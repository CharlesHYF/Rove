/*
 * 文件作用：internal/config 的单元测试 -- 验证默认值、YAML 文件加载、环境变量覆盖与非法时长报错。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/rove"
)

func TestDefaults(t *testing.T) {

	c, err := Load("")
	require.NoError(t, err)
	require.Equal(t, int64(10<<20), c.Fetch.MaxBodyBytes)
	require.Equal(t, 30*time.Second, c.Fetch.Timeout.Duration)
	require.Equal(t, 10, c.Fetch.MaxRedirects)
	require.Equal(t, 800, c.Chunk.MaxTokens)
	require.Equal(t, 100, c.Chunk.Overlap)
	require.Equal(t, "pseudo", c.Embedder.Type)
}

func TestLoadFile(t *testing.T) {

	c, err := Load(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	require.Equal(t, int64(1<<20), c.Fetch.MaxBodyBytes)
	require.Equal(t, 10*time.Second, c.Fetch.Timeout.Duration)
	require.Equal(t, 400, c.Chunk.MaxTokens)
}

func TestEnvOverride(t *testing.T) {

	t.Setenv("ROVE_ALLOW_PRIVATE", "true")
	t.Setenv("ROVE_ES_URL", "http://127.0.0.1:9200")

	c, err := Load("")
	require.NoError(t, err)
	require.True(t, c.Fetch.AllowPrivate)
	require.Equal(t, "http://127.0.0.1:9200", c.ES.URL)
}

func TestLoadInvalidDuration(t *testing.T) {

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("fetch:\n  timeout: nope\n"), 0o644))

	_, err := Load(path)

	var re *rove.Error
	require.ErrorAs(t, err, &re)
	require.Equal(t, "config.parse", re.Code)
}
