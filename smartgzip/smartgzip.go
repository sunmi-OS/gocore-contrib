package smartgzip

import (
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// GzipOnly 只对指定路径启用 gzip
func GzipOnly(paths ...string) gin.HandlerFunc {
	// 生成 map 提高匹配效率
	pathMap := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		// 确保路径以 / 开头
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		pathMap[p] = struct{}{}
	}

	// 内部调用 gin-contrib/gzip
	gz := gzip.Gzip(gzip.DefaultCompression)

	return func(c *gin.Context) {
		if _, ok := pathMap[c.FullPath()]; ok {
			// 命中目标路由才压缩
			gz(c)
		} else {
			// 否则直接放行
			c.Next()
		}
	}
}
