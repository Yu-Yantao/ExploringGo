package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

// 提取 HTML 中的纯文本内容
func ExtractPlainText(html string) (string, error) {
	// 1. 安全过滤
	p := bluemonday.UGCPolicy()
	cleanHTML := p.Sanitize(html)

	// 2. 使用 GoQuery 处理
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(cleanHTML))
	doc.Find("img").Remove() // 移除所有图片

	// 提取安全文本
	safeText := doc.Text()
	return safeText, nil
}

func main() {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>复杂富文本示例 - Elasticsearch 技术解析</title>
    <style>
        .hidden { display: none; } 
        .ad-banner { color: red; } /* 需要过滤的广告样式 */
    </style>
    <script>
        // 模拟恶意脚本
        window.addEventListener('load', () => {
            console.log('潜在 XSS 风险');
        });
    </script>
</head>
<body>
    <!-- 文章正文开始 -->
    <div class="article" id="main-content">
        <h1 data-testid="title">Elasticsearch 分布式架构解析 <span class="ad-banner">[广告]</span></h1>
        
        <div class="author-info">
            <p>作者：<a href="javascript:alert('xss')" onclick="track()">技术达人</a> | 发布日期：2023-10-01</p>
        </div>

        <section class="content">
            <p>Elasticsearch 基于 <strong>Lucene</strong> 构建，采用 <em>分布式设计</em>，其核心特性包括：</p>
            
            <ul>
                <li>水平扩展：通过 <code style="color: blue;">shard</code> 分片实现数据分布</li>
                <li>高可用：<span class="hidden">敏感信息</span>副本机制（Replica）</li>
                <li>近实时搜索：<b>refresh_interval</b> 控制可见性</li>
            </ul>

            <pre><code class="language-javascript">
// 示例：Elasticsearch 索引文档
client.index({
    index: 'articles',
    body: {
        title: "Hello World",
        content: "分布式搜索实践"
    }
});
            </code></pre>

            <div class="chart-container">
                <img src="data:image/png;base64,..." alt="架构图：节点与分片关系" 
                     onerror="alert('图片加载失败')" />
                <p style="font-size: 0.9em;">图 1：<i>集群节点拓扑</i></p>
            </div>

            <table border="1">
                <thead>
                    <tr><th>参数</th><th>默认值</th></tr>
                </thead>
                <tbody>
                    <tr><td>number_of_shards</td><td>5</td></tr>
                    <tr><td>number_of_replicas</td><td>1</td></tr>
                </tbody>
            </table>

            <iframe width="560" height="315" src="https://www.youtube.com/embed/abc123" 
                    frameborder="0" allow="accelerometer; encrypted-media; gyroscope"></iframe>

            <div class="advertisement">
                <script>document.write("动态广告内容");</script>
            </div>
        </section>
    </div>
    <!-- 文章正文结束 -->

    <noscript>
        <p>请启用 JavaScript 以获得完整体验</p>
    </noscript>
</body>
</html>`
	plainText, err := ExtractPlainText(html)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(strings.ReplaceAll(plainText, "\n", ""))
}
