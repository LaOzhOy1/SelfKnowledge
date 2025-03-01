package main

import (
	"fmt"
	"os"
)

type point struct {
	x int
	y int
}

var dirs = [4]point{
	{-1, 0}, {0, -1}, {1, 0}, {0, 1},
}

func (p point) add(r point) point {
	return point{p.x + r.x, p.y + r.y}
}

func (p point) at(matrix [][]int) (int, bool) {
	if p.x < 0 || p.x >= len(matrix) || p.y < 0 || p.y >= len(matrix[p.x]) {
		return 0, true
	}
	return matrix[p.x][p.y], false
}

func walk(matrix [][]int, start point, end point) [][]int {
	steps := make([][]int, len(matrix))
	for i := 0; i < len(steps); i++ {
		steps[i] = make([]int, len(matrix[i]))
	}

	pointArrival := []point{start}

	for len(pointArrival) > 0 {
		cur := pointArrival[0]
		pointArrival = pointArrival[1:]
		if cur == end {
			return steps
		}
		for _, dir := range dirs {
			next := cur.add(dir)
			if value, isOut := next.at(matrix); isOut == true || value == 1 {
				continue
			}
			value, _ := next.at(steps)
			if value != 0 {
				continue
			}
			if start == next {
				continue
			}
			step, _ := cur.at(steps)
			steps[next.x][next.y] = step + 1
			pointArrival = append(pointArrival, next)
		}
	}
	return steps
}

func readMaze(filename string) [][]int {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	var row, col int
	_, _ = fmt.Fscanf(file, "%d %d", &row, &col)
	maze := make([][]int, row)
	for i := 0; i < row; i++ {
		maze[i] = make([]int, col)
		for j := 0; j < col; j++ {
			fmt.Fscanf(file, "%d", &maze[i][j])
		}
	}
	return maze
}

func main() {
	maze := readMaze("leecode/bfs/maze.txt")
	for _, row := range maze {
		for _, val := range row {
			fmt.Printf("%d ", val)
		}
		fmt.Println()
	}

	steps := walk(maze, point{0, 0}, point{
		x: 5,
		y: 4,
	})
	for _, row := range steps {
		for _, val := range row {
			fmt.Printf("%d ", val)
		}
		fmt.Println()
	}

}
