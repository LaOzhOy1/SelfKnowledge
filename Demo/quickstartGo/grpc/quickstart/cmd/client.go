/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"quickstart/config"
	"quickstart/proto"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("send hello request")
		// 生成链接
		dial, err := grpc.Dial(config.NewConfig().GrpcServerAddr, grpc.WithInsecure())
		defer func(dial *grpc.ClientConn) {
			err2 := dial.Close()
			if err2 != nil {
				fmt.Println("dial close fail!")
				return
			}
		}(dial)
		if err != nil {
			fmt.Println("dial fail!")
			return
		}

		// 配置客户端链接
		client := proto.NewHelloServiceClient(dial)
		resp, err := client.SayHello(context.Background(), &proto.HelloRequest{Pid: 1})
		if err != nil {
			fmt.Println("resp fail!")
			return
		}
		fmt.Printf("get resp is %d\n", resp.Result)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
