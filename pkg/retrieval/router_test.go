/*
 * 文件作用：Query Router 单元测试 -- 垂类推断优先级与参数校验（PRD FR-QRY-001）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package retrieval

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouteVerticalCode(t *testing.T) {

	require.Equal(t, VerticalCode, RouteVertical("func main() { fmt.Println }"))
	require.Equal(t, VerticalCode, RouteVertical("how to pip install rove"))
	require.Equal(t, VerticalCode, RouteVertical("https://github.com/modelcontextprotocol/go-sdk snippet"))
}

func TestRouteVerticalAcademic(t *testing.T) {

	require.Equal(t, VerticalAcademic, RouteVertical("arxiv paper attention is all you need"))
	require.Equal(t, VerticalAcademic, RouteVertical("citation doi.org/10.1000 test"))
}

func TestRouteVerticalDocs(t *testing.T) {

	require.Equal(t, VerticalDocs, RouteVertical("how to install nginx"))
	require.Equal(t, VerticalDocs, RouteVertical("usage guide for configuration"))
}

func TestRouteVerticalWeb(t *testing.T) {

	require.Equal(t, VerticalWeb, RouteVertical("latest news today"))
	require.Equal(t, VerticalWeb, RouteVertical("browser agents"))
}

func TestParseVertical(t *testing.T) {

	for _, value := range []string{"", "auto"} {
		vertical, err := ParseVertical(value)
		require.NoError(t, err)
		require.Equal(t, VerticalAuto, vertical)
	}
	vertical, err := ParseVertical("WEB")
	require.NoError(t, err)
	require.Equal(t, VerticalWeb, vertical)

	_, err = ParseVertical("image")
	require.Error(t, err, "非法垂类应报错")
}
