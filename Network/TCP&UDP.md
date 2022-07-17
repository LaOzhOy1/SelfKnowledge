# TCP





## TCP 握手



### 服务端初始化



#### 服务端通过listen扫描寻找可用端口进行监听

- 半链接队列初始化
- 全连接队列初始化
- 将Socket中的Sock状态从SS_UNCONNECTED设置为TCP_LISTEN

```c

/*
 *	Perform a listen. Basically, we allow the protocol to do anything
 *	necessary for a listen, and if that works, we mark the socket as
 *	ready for listening.
 */

int __sys_listen(int fd, int backlog)
{
	struct socket *sock;
	int err, fput_needed;
	int somaxconn;

	sock = sockfd_lookup_light(fd, &err, &fput_needed);
	if (sock) {
		somaxconn = sock_net(sock->sk)->core.sysctl_somaxconn;
		if ((unsigned int)backlog > somaxconn)
			backlog = somaxconn;

		err = security_socket_listen(sock, backlog);
		if (!err)
			err = sock->ops->listen(sock, backlog);

		fput_light(sock->file, fput_needed);
	}
	return err;
}
```
```C
/*
 *	Move a socket into listening state.
 */
int inet_listen(struct socket *sock, int backlog)
{
	struct sock *sk = sock->sk;
	unsigned char old_state;
	int err, tcp_fastopen;

	lock_sock(sk);

	err = -EINVAL;
	if (sock->state != SS_UNCONNECTED || sock->type != SOCK_STREAM)
		goto out;

	old_state = sk->sk_state;
	if (!((1 << old_state) & (TCPF_CLOSE | TCPF_LISTEN)))
		goto out;

	WRITE_ONCE(sk->sk_max_ack_backlog, backlog);
	/* Really, if the socket is already in listen state
	 * we can only allow the backlog to be adjusted.
	 */
	if (old_state != TCP_LISTEN) {
		/* Enable TFO w/o requiring TCP_FASTOPEN socket option.
		 * Note that only TCP sockets (SOCK_STREAM) will reach here.
		 * Also fastopen backlog may already been set via the option
		 * because the socket was in TCP_LISTEN state previously but
		 * was shutdown() rather than close().
		 */
		tcp_fastopen = sock_net(sk)->ipv4.sysctl_tcp_fastopen;
		if ((tcp_fastopen & TFO_SERVER_WO_SOCKOPT1) &&
		    (tcp_fastopen & TFO_SERVER_ENABLE) &&
		    !inet_csk(sk)->icsk_accept_queue.fastopenq.max_qlen) {
			fastopen_queue_tune(sk, backlog);
			tcp_fastopen_init_key_once(sock_net(sk));
		}

		err = inet_csk_listen_start(sk);
		if (err)
			goto out;
		tcp_call_bpf(sk, BPF_SOCK_OPS_TCP_LISTEN_CB, 0, NULL);
	}
	err = 0;

out:
	release_sock(sk);
	return err;
}
```

```C

int inet_csk_listen_start(struct sock *sk)
{
	struct inet_connection_sock *icsk = inet_csk(sk);
	struct inet_sock *inet = inet_sk(sk);
	int err = -EADDRINUSE;

	reqsk_queue_alloc(&icsk->icsk_accept_queue);

	sk->sk_ack_backlog = 0;
	inet_csk_delack_init(sk);

	if (sk->sk_txrehash == SOCK_TXREHASH_DEFAULT)
		sk->sk_txrehash = READ_ONCE(sock_net(sk)->core.sysctl_txrehash);

	/* There is race window here: we announce ourselves listening,
	 * but this transition is still not validated by get_port().
	 * It is OK, because this socket enters to hash table only
	 * after validation is complete.
	 */
	inet_sk_state_store(sk, TCP_LISTEN);
	if (!sk->sk_prot->get_port(sk, inet->inet_num)) {
		inet->inet_sport = htons(inet->inet_num);

		sk_dst_reset(sk);
		err = sk->sk_prot->hash(sk);

		if (likely(!err))
			return 0;
	}

	inet_sk_set_state(sk, TCP_CLOSE);
	return err;
}
```

##### 全连接队列初始化

```c

void reqsk_queue_alloc(struct request_sock_queue *queue)
{
	spin_lock_init(&queue->rskq_lock);

	spin_lock_init(&queue->fastopenq.lock);
	queue->fastopenq.rskq_rst_head = NULL;
	queue->fastopenq.rskq_rst_tail = NULL;
	queue->fastopenq.qlen = 0;

	queue->rskq_accept_head = NULL;
}

```









### 客户端链接&三次握手过程



#### 第一次握手

-  对于AF_INET4 TCP的Socket来说，Connect 调用的是tcp_v4_connect, 设置当前Socket状态为TCP_SYN_SENT 并发送SYN报文
- 完成tcp_v4_connect 后，Socket状态从SS_UNCONNECTED转到TCP_SYN_SENT 再到  SS_CONNECTING

```C
int __inet_stream_connect(struct socket *sock, struct sockaddr *uaddr,
			  int addr_len, int flags, int is_sendmsg){
               
    ....
    
	case SS_UNCONNECTED:
		err = -EISCONN;
		if (sk->sk_state != TCP_CLOSE)
			goto out;

		if (BPF_CGROUP_PRE_CONNECT_ENABLED(sk)) {
			err = sk->sk_prot->pre_connect(sk, uaddr, addr_len);
			if (err)
				goto out;
		}

		err = sk->sk_prot->connect(sk, uaddr, addr_len);
		if (err < 0)
			goto out;

		sock->state = SS_CONNECTING;

		if (!err && inet_sk(sk)->defer_connect)
			goto out;

		/* Just entered SS_CONNECTING state; the only
		 * difference is that return value in non-blocking
		 * case is EINPROGRESS, rather than EALREADY.
		 */
		err = -EINPROGRESS;
		break;
	}
	
	.....
	
}
```

#####  选择可用端口

```c
/* This will initiate an outgoing connection. */
int tcp_v4_connect(struct sock *sk, struct sockaddr *uaddr, int addr_len)
{
    .......
    
	/* Socket identity is still unknown (sport may be zero).
	 * However we set state to SYN-SENT and not releasing socket
	 * lock select source port, enter ourselves into the hash tables and
	 * complete initialization after this.
	 */
	tcp_set_state(sk, TCP_SYN_SENT);
	err = inet_hash_connect(tcp_death_row, sk);
	if (err)
		goto failure;

	... 
	err = tcp_connect(sk);

	if (err)
		goto failure;

	return 0;

    ... 
}
```

- 生成端口扫描随机数

```C
/*
 * Bind a port for a connect operation and hash it.
 */
int inet_hash_connect(struct inet_timewait_death_row *death_row,
		      struct sock *sk)
{
	u64 port_offset = 0;

	if (!inet_sk(sk)->inet_num)
		port_offset = inet_sk_port_offset(sk);
	return __inet_hash_connect(death_row, sk, port_offset,
				   __inet_check_established);
}
```

- 循环检测范围内的端口是否可用
  - 如果有指定端口的话，直接通过check_established检测端口是否进入链接状态
  - 通过bhash 记录使用过的端口
  - 如果bhash记录有其他链接使用过这个端口，但是check_established检查端口中要建立的链接四元组信息是唯一的，则可以创建一条新的链接

```c

int __inet_hash_connect(struct inet_timewait_death_row *death_row,
		struct sock *sk, u64 port_offset,
		int (*check_established)(struct inet_timewait_death_row *,
			struct sock *, __u16, struct inet_timewait_sock **))
{

	int port = inet_sk(sk)->inet_num;

	if (port) {
		head = &hinfo->bhash[inet_bhashfn(net, port,
						  hinfo->bhash_size)];
		tb = inet_csk(sk)->icsk_bind_hash;
		spin_lock_bh(&head->lock);
		if (sk_head(&tb->owners) == sk && !sk->sk_bind_node.next) {
			inet_ehash_nolisten(sk, NULL, NULL);
			spin_unlock_bh(&head->lock);
			return 0;
		}
		spin_unlock(&head->lock);
		/* No definite answer... Walk to established hash table */
		ret = check_established(death_row, sk, port, NULL);
		local_bh_enable();
		return ret;
	}


	inet_get_local_port_range(net, &low, &high);
	high++; /* [32768, 60999] -> [32768, 61000[ */
	remaining = high - low;
	if (likely(remaining > 1))
		remaining &= ~1U;

	net_get_random_once(table_perturb,
			    INET_TABLE_PERTURB_SIZE * sizeof(*table_perturb));
	index = port_offset & (INET_TABLE_PERTURB_SIZE - 1);

	offset = READ_ONCE(table_perturb[index]) + (port_offset >> 32);
	offset %= remaining;

	/* In first pass we try ports of @low parity.
	 * inet_csk_get_port() does the opposite choice.
	 */
	offset &= ~1U;
other_parity_scan:
	port = low + offset;
	for (i = 0; i < remaining; i += 2, port += 2) {
		if (unlikely(port >= high))
			port -= remaining;
		if (inet_is_local_reserved_port(net, port))
			continue;
		head = &hinfo->bhash[inet_bhashfn(net, port,
						  hinfo->bhash_size)];
		spin_lock_bh(&head->lock);

		/* Does not bother with rcv_saddr checks, because
		 * the established check is already unique enough.
		 */
		inet_bind_bucket_for_each(tb, &head->chain) {
			if (net_eq(ib_net(tb), net) && tb->l3mdev == l3mdev &&
			    tb->port == port) {
				if (tb->fastreuse >= 0 ||
				    tb->fastreuseport >= 0)
					goto next_port;
				WARN_ON(hlist_empty(&tb->owners));
				if (!check_established(death_row, sk,
						       port, &tw))
					goto ok;
				goto next_port;
			}
		}

		goto ok;
next_port:
		spin_unlock_bh(&head->lock);
		cond_resched();
	}

	offset++;
	if ((offset & 1) && remaining > 1)
		goto other_parity_scan;

	return -EADDRNOTAVAIL;

ok:
	/* Here we want to add a little bit of randomness to the next source
	 * port that will be chosen. We use a max() with a random here so that
	 * on low contention the randomness is maximal and on high contention
	 * it may be inexistent.
	 */
	i = max_t(int, i, (prandom_u32() & 7) * 2);
	WRITE_ONCE(table_perturb[index], READ_ONCE(table_perturb[index]) + i + 2);

	/* Head lock still held and bh's disabled */
	inet_bind_hash(sk, tb, port);
	if (sk_unhashed(sk)) {
		inet_sk(sk)->inet_sport = htons(port);
		inet_ehash_nolisten(sk, (struct sock *)tw, NULL);
	}
	if (tw)
		inet_twsk_bind_unhash(tw, hinfo);
	spin_unlock(&head->lock);
	if (tw)
		inet_twsk_deschedule_put(tw);
	local_bh_enable();
	return 0;
}

```

- 当Socket要建立链接的四元组信息和现有的TCP链接四元组信息完全一致则端口不可用，否则返回0，端口可用

```C
/* called with local bh disabled */
static int __inet_check_established(struct inet_timewait_death_row *death_row,
				    struct sock *sk, __u16 lport,
				    struct inet_timewait_sock **twp)
{

	struct inet_ehash_bucket *head = inet_ehash_bucket(hinfo, hash);


	spin_lock(lock);

	sk_nulls_for_each(sk2, node, &head->chain) {
		if (sk2->sk_hash != hash)
			continue;

		if (likely(inet_match(net, sk2, acookie, ports, dif, sdif))) {
			if (sk2->sk_state == TCP_TIME_WAIT) {
				tw = inet_twsk(sk2);
				if (twsk_unique(sk, sk2, twp))
					break;
			}
			goto not_unique;
		}
	}

	return 0;

not_unique:
	spin_unlock(lock);
	return -EADDRNOTAVAIL;
}

```

##### 发起Syn请求

- 申请并初始化SKB
- 将SYN报文写到发送队列中
- 将SYN报文通过transmit 方法传送出去
- 启动重传计时器

```c
/* Build a SYN and send it off. */
int tcp_connect(struct sock *sk)
{
	struct tcp_sock *tp = tcp_sk(sk);
	struct sk_buff *buff;
	int err;

	tcp_call_bpf(sk, BPF_SOCK_OPS_TCP_CONNECT_CB, 0, NULL);

	if (inet_csk(sk)->icsk_af_ops->rebuild_header(sk))
		return -EHOSTUNREACH; /* Routing failure or similar. */

	tcp_connect_init(sk);

	if (unlikely(tp->repair)) {
		tcp_finish_connect(sk, NULL);
		return 0;
	}

	buff = tcp_stream_alloc_skb(sk, 0, sk->sk_allocation, true);
	if (unlikely(!buff))
		return -ENOBUFS;

    // 将当前seq + 1 写入到 http报文中
	tcp_init_nondata_skb(buff, tp->write_seq++, TCPHDR_SYN);
	tcp_mstamp_refresh(tp);
	tp->retrans_stamp = tcp_time_stamp(tp);
    // 将Syn报文写入发送队列中
	tcp_connect_queue_skb(sk, buff);
	tcp_ecn_send_syn(sk, buff);
	tcp_rbtree_insert(&sk->tcp_rtx_queue, buff);

	/* Send off SYN; include data in Fast Open. */
    // 实际发送SYN报文
	err = tp->fastopen_req ? tcp_send_syn_data(sk, buff) :
	      tcp_transmit_skb(sk, buff, 1, sk->sk_allocation);
	if (err == -ECONNREFUSED)
		return err;

	/* We change tp->snd_nxt after the tcp_transmit_skb() call
	 * in order to make this packet get counted in tcpOutSegs.
	 */
	WRITE_ONCE(tp->snd_nxt, tp->write_seq);
	tp->pushed_seq = tp->write_seq;
	buff = tcp_send_head(sk);
	if (unlikely(buff)) {
		WRITE_ONCE(tp->snd_nxt, TCP_SKB_CB(buff)->seq);
		tp->pushed_seq	= TCP_SKB_CB(buff)->seq;
	}
	TCP_INC_STATS(sock_net(sk), TCP_MIB_ACTIVEOPENS);

	/* Timer for repeating the SYN until an answer. */
    // 启动重传计时器
	inet_csk_reset_xmit_timer(sk, ICSK_TIME_RETRANS,
				  inet_csk(sk)->icsk_rto, TCP_RTO_MAX);
	return 0;
}

```



#### 第二次握手



##### Listen状态端口进行响应

- 通过tcp_v4_cookie_check 查找半链接队列中的请求信息，对于第一次握手的请求，则查找为空
- 通过调用tcp_rcv_state_process 处理SYN请求

```C
/* The socket must have it's spinlock held when we get
 * here, unless it is a TCP_LISTEN socket.
 *
 * We have a potential double-lock case here, so even when
 * doing backlog processing we use the BH locking scheme.
 * This is because we cannot sleep with the original spinlock
 * held.
 */
int tcp_v4_do_rcv(struct sock *sk, struct sk_buff *skb)
{
	enum skb_drop_reason reason;
	struct sock *rsk;

	reason = SKB_DROP_REASON_NOT_SPECIFIED;
	if (tcp_checksum_complete(skb))
		goto csum_err;

	if (sk->sk_state == TCP_LISTEN) {
        
		struct sock *nsk = tcp_v4_cookie_check(sk, skb);

		if (!nsk)
			goto discard;
		if (nsk != sk) {
			if (tcp_child_process(sk, nsk, skb)) {
				rsk = nsk;
				goto reset;
			}
			return 0;
		}
	} else
		sock_rps_save_rxhash(sk, skb);

    // 处理SYN请求
	if (tcp_rcv_state_process(sk, skb)) {
		rsk = sk;
		goto reset;
	}
	return 0;

reset:
	tcp_v4_send_reset(rsk, skb);
discard:
	kfree_skb_reason(skb, reason);
	/* Be careful here. If this function gets more complicated and
	 * gcc suffers from register pressure on the x86, sk (in %ebx)
	 * might be destroyed here. This current version compiles correctly,
	 * but you have been warned.
	 */
	return 0;

csum_err:
	reason = SKB_DROP_REASON_TCP_CSUM;
	trace_tcp_bad_csum(skb);
	TCP_INC_STATS(sock_net(sk), TCP_MIB_CSUMERRORS);
	TCP_INC_STATS(sock_net(sk), TCP_MIB_INERRS);
	goto discard;
}
```

 ##### 响应SYN请求

- 当队列满时，对请求直接抛弃

```C
int tcp_rcv_state_process(struct sock *sk, struct sk_buff *skb)
{
...
	case TCP_LISTEN:
		if (th->ack)
			return 1;

		if (th->rst) {
			SKB_DR_SET(reason, TCP_RESET);
			goto discard;
		}
		if (th->syn) {
			if (th->fin) {
				SKB_DR_SET(reason, TCP_FLAGS);
				goto discard;
			}
			/* It is possible that we process SYN packets from backlog,
			 * so we need to make sure to disable BH and RCU right there.
			 */
			rcu_read_lock();
			local_bh_disable();
			acceptable = icsk->icsk_af_ops->conn_request(sk, skb) >= 0;
			local_bh_enable();
			rcu_read_unlock();

			if (!acceptable)
				return 1;
			consume_skb(skb);
			return 0;
		}
		SKB_DR_SET(reason, TCP_FLAGS);
		goto discard;
...
}


int tcp_conn_request(struct request_sock_ops *rsk_ops,
		     const struct tcp_request_sock_ops *af_ops,
		     struct sock *sk, struct sk_buff *skb)
{


	/* TW buckets are converted to open requests without
	 * limitations, they conserve resources and peer is
	 * evidently real one.
	 */
    //  检查半链接队列是否为满
	if ((net->ipv4.sysctl_tcp_syncookies == 2 ||
	     inet_csk_reqsk_queue_is_full(sk)) && !isn) {
		want_cookie = tcp_syn_flood_action(sk, rsk_ops->slab_name);
		if (!want_cookie)
			goto drop;
	}

    // 检查全连接队列是否为满
	if (sk_acceptq_is_full(sk)) {
		NET_INC_STATS(sock_net(sk), LINUX_MIB_LISTENOVERFLOWS);
		goto drop;
	}

	req = inet_reqsk_alloc(rsk_ops, sk, !want_cookie);
	if (!req)
		goto drop;

	req->syncookie = want_cookie;
	tcp_rsk(req)->af_specific = af_ops;
	tcp_rsk(req)->ts_off = 0;
#if IS_ENABLED(CONFIG_MPTCP)
	tcp_rsk(req)->is_mptcp = 0;
#endif

	tcp_clear_options(&tmp_opt);
	tmp_opt.mss_clamp = af_ops->mss_clamp;
	tmp_opt.user_mss  = tp->rx_opt.user_mss;
	tcp_parse_options(sock_net(sk), skb, &tmp_opt, 0,
			  want_cookie ? NULL : &foc);

	if (want_cookie && !tmp_opt.saw_tstamp)
		tcp_clear_options(&tmp_opt);

	if (IS_ENABLED(CONFIG_SMC) && want_cookie)
		tmp_opt.smc_ok = 0;

	tmp_opt.tstamp_ok = tmp_opt.saw_tstamp;
	tcp_openreq_init(req, &tmp_opt, skb, sk);
	inet_rsk(req)->no_srccheck = inet_sk(sk)->transparent;

	/* Note: tcp_v6_init_req() might override ir_iif for link locals */
	inet_rsk(req)->ir_iif = inet_request_bound_dev_if(sk, skb);

	dst = af_ops->route_req(sk, skb, &fl, req);
	if (!dst)
		goto drop_and_free;

	if (tmp_opt.tstamp_ok)
		tcp_rsk(req)->ts_off = af_ops->init_ts_off(net, skb);

	if (!want_cookie && !isn) {
		/* Kill the following clause, if you dislike this way. */
		if (!net->ipv4.sysctl_tcp_syncookies &&
		    (net->ipv4.sysctl_max_syn_backlog - inet_csk_reqsk_queue_len(sk) <
		     (net->ipv4.sysctl_max_syn_backlog >> 2)) &&
		    !tcp_peer_is_proven(req, dst)) {
			/* Without syncookies last quarter of
			 * backlog is filled with destinations,
			 * proven to be alive.
			 * It means that we continue to communicate
			 * to destinations, already remembered
			 * to the moment of synflood.
			 */
			pr_drop_req(req, ntohs(tcp_hdr(skb)->source),
				    rsk_ops->family);
			goto drop_and_release;
		}

		isn = af_ops->init_seq(skb);
	}

	tcp_ecn_create_request(req, skb, sk, dst);

	if (want_cookie) {
		isn = cookie_init_sequence(af_ops, sk, skb, &req->mss);
		if (!tmp_opt.tstamp_ok)
			inet_rsk(req)->ecn_ok = 0;
	}

	tcp_rsk(req)->snt_isn = isn;
	tcp_rsk(req)->txhash = net_tx_rndhash();
	tcp_rsk(req)->syn_tos = TCP_SKB_CB(skb)->ip_dsfield;
	tcp_openreq_init_rwin(req, sk, dst);
	sk_rx_queue_set(req_to_sk(req), skb);
	if (!want_cookie) {
		tcp_reqsk_record_syn(sk, req, skb);
		fastopen_sk = tcp_try_fastopen(sk, skb, req, &foc, dst);
	}
	if (fastopen_sk) {
		af_ops->send_synack(fastopen_sk, dst, &fl, req,
				    &foc, TCP_SYNACK_FASTOPEN, skb);
		/* Add the child socket directly into the accept queue */
		if (!inet_csk_reqsk_queue_add(sk, req, fastopen_sk)) {
			reqsk_fastopen_remove(fastopen_sk, req, false);
			bh_unlock_sock(fastopen_sk);
			sock_put(fastopen_sk);
			goto drop_and_free;
		}
		sk->sk_data_ready(sk);
		bh_unlock_sock(fastopen_sk);
		sock_put(fastopen_sk);
	} else {
		tcp_rsk(req)->tfo_listener = false;
		if (!want_cookie) {
			req->timeout = tcp_timeout_init((struct sock *)req);
            // 将请求放入半连接队列中，并开启计时器
			inet_csk_reqsk_queue_hash_add(sk, req, req->timeout);
		}
        // 发送ack + seq 确认包
		af_ops->send_synack(sk, dst, &fl, req, &foc,
				    !want_cookie ? TCP_SYNACK_NORMAL :
						   TCP_SYNACK_COOKIE,
				    skb);
		if (want_cookie) {
			reqsk_free(req);
			return 0;
		}
	}
	reqsk_put(req);
	return 0;

drop_and_release:
	dst_release(dst);
drop_and_free:
	__reqsk_free(req);
drop:
	tcp_listendrop(sk);
	return 0;
}
```

#### 第三次握手

##### 客户端接收服务端的响应

- 删除发送队列
- 删除定时器
- 开启保活定时器
- 初始化拥塞控制
- 修改Socket为Established
- 调用 tcp_send_ack 发送ACK包

```c
int tcp_rcv_state_process(struct sock *sk, struct sk_buff *skb)
{
.....
	switch (sk->sk_state) {

	case TCP_SYN_SENT:
		tp->rx_opt.saw_tstamp = 0;
		tcp_mstamp_refresh(tp);
		queued = tcp_rcv_synsent_state_process(sk, skb, th);
		if (queued >= 0)
			return queued;

		/* Do step6 onward by hand. */
		tcp_urg(sk, skb, th);
		__kfree_skb(skb);
		tcp_data_snd_check(sk);
		return 0;
	}
....
}

static int tcp_rcv_synsent_state_process(struct sock *sk, struct sk_buff *skb,
					 const struct tcphdr *th)
{

		if (th->ack) {
		/* rfc793:
		 * "If the state is SYN-SENT then
		 *    first check the ACK bit
		 *      If the ACK bit is set
		 *	  If SEG.ACK =< ISS, or SEG.ACK > SND.NXT, send
		 *        a reset (unless the RST bit is set, if so drop
		 *        the segment and return)"
		 */
		if (!after(TCP_SKB_CB(skb)->ack_seq, tp->snd_una) ||
		    after(TCP_SKB_CB(skb)->ack_seq, tp->snd_nxt)) {
			/* Previous FIN/ACK or RST/ACK might be ignored. */
			if (icsk->icsk_retransmits == 0)
				inet_csk_reset_xmit_timer(sk,
						ICSK_TIME_RETRANS,
						TCP_TIMEOUT_MIN, TCP_RTO_MAX);
			goto reset_and_undo;
		}

		if (tp->rx_opt.saw_tstamp && tp->rx_opt.rcv_tsecr &&
		    !between(tp->rx_opt.rcv_tsecr, tp->retrans_stamp,
			     tcp_time_stamp(tp))) {
			NET_INC_STATS(sock_net(sk),
					LINUX_MIB_PAWSACTIVEREJECTED);
			goto reset_and_undo;
		}

		/* Now ACK is acceptable.
		 *
		 * "If the RST bit is set
		 *    If the ACK was acceptable then signal the user "error:
		 *    connection reset", drop the segment, enter CLOSED state,
		 *    delete TCB, and return."
		 */

		tcp_ack(sk, skb, FLAG_SLOWPATH);
.....

		tcp_finish_connect(sk, skb);

	...
		if (sk->sk_write_pending ||
		    icsk->icsk_accept_queue.rskq_defer_accept ||
		    inet_csk_in_pingpong_mode(sk)) {
			/* Save one ACK. Data will be ready after
			 * several ticks, if write_pending is set.
			 *
			 * It may be deleted, but with this feature tcpdumps
			 * look so _wonderfully_ clever, that I was not able
			 * to stand against the temptation 8)     --ANK
			 */
			inet_csk_schedule_ack(sk);
			tcp_enter_quickack_mode(sk, TCP_MAX_QUICKACKS);
			inet_csk_reset_xmit_timer(sk, ICSK_TIME_DACK,
						  TCP_DELACK_MAX, TCP_RTO_MAX);
			goto consume;
		}
           
		tcp_send_ack(sk);
		return -1;
	}
}
```

##### 服务端响应客户端ACK

- 修改当前Socket状态为Established
- 判断全连接队列中请求是否已经满了，如果满了则直接丢弃
- 检查ACK请求是否在半连接队列中，如果存在，则创建子Socket，将其放入全连接队列后删除在半链接队列中的位置

```C
int tcp_v4_do_rcv(struct sock *sk, struct sk_buff *skb)
{
	enum skb_drop_reason reason;
	struct sock *rsk;

	reason = SKB_DROP_REASON_NOT_SPECIFIED;
	if (tcp_checksum_complete(skb))
		goto csum_err;

	if (sk->sk_state == TCP_LISTEN) {
        
		struct sock *nsk = tcp_v4_cookie_check(sk, skb);

		if (!nsk)
			goto discard;
		if (nsk != sk) {
			if (tcp_child_process(sk, nsk, skb)) {
				rsk = nsk;
				goto reset;
			}
			return 0;
		}
	} else
		sock_rps_save_rxhash(sk, skb);

	return 0;
}
```

```
int tcp_rcv_state_process(struct sock *sk, struct sk_buff *skb)
{

	case TCP_SYN_RECV:
	.... 设置当前Socket 状态
		tcp_set_state(sk, TCP_ESTABLISHED);
		sk->sk_state_change(sk);
		...
		break;
}
```





##### Accept 接收消息

- 从全连接队列头部取出一个请求进行处理
- 如果时间没有超过保活计时器的时间，读取请求后不会直接释放Sock内存，而是将他置为空

```C
/*
 * This will accept the next outstanding connection.
 */
struct sock *inet_csk_accept(struct sock *sk, int flags, int *err, bool kern)
{
	
	/* Find already established connection */
	if (reqsk_queue_empty(queue)) {
		long timeo = sock_rcvtimeo(sk, flags & O_NONBLOCK);

		/* If this is a non blocking socket don't sleep */
		error = -EAGAIN;
		if (!timeo)
			goto out_err;

		error = inet_csk_wait_for_connect(sk, timeo);
		if (error)
			goto out_err;
	}
	req = reqsk_queue_remove(queue, sk);
	newsk = req->sk;

	if (sk->sk_protocol == IPPROTO_TCP &&
	    tcp_rsk(req)->tfo_listener) {
		spin_lock_bh(&queue->fastopenq.lock);
		if (tcp_rsk(req)->tfo_listener) {
			/* We are still waiting for the final ACK from 3WHS
			 * so can't free req now. Instead, we set req->sk to
			 * NULL to signify that the child socket is taken
			 * so reqsk_fastopen_remove() will free the req
			 * when 3WHS finishes (or is aborted).
			 */
			req->sk = NULL;
			req = NULL;
		}
		spin_unlock_bh(&queue->fastopenq.lock);
	}

out:
	release_sock(sk);
	if (newsk && mem_cgroup_sockets_enabled) {
		int amt;

		/* atomically get the memory usage, set and charge the
		 * newsk->sk_memcg.
		 */
		lock_sock(newsk);

		/* The socket has not been accepted yet, no need to look at
		 * newsk->sk_wmem_queued.
		 */
		amt = sk_mem_pages(newsk->sk_forward_alloc +
				   atomic_read(&newsk->sk_rmem_alloc));
		mem_cgroup_sk_alloc(newsk);
		if (newsk->sk_memcg && amt)
			mem_cgroup_charge_skmem(newsk->sk_memcg, amt,
						GFP_KERNEL | __GFP_NOFAIL);

		release_sock(newsk);
	}
	if (req)
		reqsk_put(req);
	return newsk;
out_err:
	newsk = NULL;
	req = NULL;
	*err = error;
	goto out;
}
```



### 总结









## TCP 挥手





















