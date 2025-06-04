//go:build !solution

package hogwarts

type Color byte

const (
	WHITE Color = iota
	GRAY
	BLACK
)

func dfs(node string, graph map[string][]string, color map[string]Color, result *[]string) {
	callStack := []string{node}
	for len(callStack) > 0 {
		curNode := callStack[len(callStack)-1]
		if color[curNode] == GRAY {
			callStack = callStack[:len(callStack)-1]
			color[curNode] = BLACK
			*result = append(*result, curNode)
			continue
		}
		color[curNode] = GRAY
		for _, neighbour := range graph[curNode] {
			if color[neighbour] == GRAY {
				panic("loop in the graph")
			}
			if color[neighbour] == WHITE {
				callStack = append(callStack, neighbour)
			}
		}
	}
}

func GetCourseList(prereqs map[string][]string) []string {
	graph := make(map[string][]string)
	for node, fromNodes := range prereqs {
		for _, fromNode := range fromNodes {
			graph[fromNode] = append(graph[fromNode], node)
		}
	}

	color := make(map[string]Color)
	result := make([]string, 0)
	for node := range graph {
		if color[node] == WHITE {
			dfs(node, graph, color, &result)
		}
	}

	reversed := make([]string, 0)
	for i := len(result) - 1; i >= 0; i-- {
		reversed = append(reversed, result[i])
	}
	return reversed
}
