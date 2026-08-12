/*
 * 文件作用：PageState -- 交互元素与稳定 element id（DOM CSS 路径）输出（PRD FR-BRW-002）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package browser

// InteractiveElement 是页面上的可交互元素（ID 为 DOM 路径，结构稳定时跨加载稳定）。
type InteractiveElement struct {
	ID   string
	Tag  string
	Text string
	Type string
	Href string
}

// PageState 是浏览快照。
type PageState struct {
	URL         string
	Title       string
	Text        string
	Links       []string
	Interactive []InteractiveElement
}

// pageStateJS 单次求值返回文本/链接/交互元素（含 DOM 路径 id）。
// id 语义：html > body > main > a:nth-of-type(2) 形式的确定性路径；
// DOM 结构不变时两次加载 id 一致；动态 DOM 变化时路径随之变化（降级，文档明示）。
const pageStateJS = `(() => {
	const domPath = (el) => {
		const parts = [];
		let node = el;
		while (node && node.nodeType === 1) {
			let sel = node.tagName.toLowerCase();
			const parent = node.parentElement;
			if (parent) {
				const siblings = Array.from(parent.children).filter((c) => c.tagName === node.tagName);
				if (siblings.length > 1) {
					sel += ':nth-of-type(' + (siblings.indexOf(node) + 1) + ')';
				}
			}
			parts.unshift(sel);
			node = node.parentElement;
		}
		return parts.join(' > ');
	};
	const selector = 'a, button, input, select, textarea, [role="button"], [role="link"]';
	return {
		text: document.body ? document.body.innerText.slice(0, 20000) : '',
		links: Array.from(document.querySelectorAll('a[href]')).map((a) => a.href).slice(0, 500),
		interactive: Array.from(document.querySelectorAll(selector)).map((el) => ({
			id: domPath(el),
			tag: el.tagName.toLowerCase(),
			text: (el.innerText || el.value || '').trim().slice(0, 200),
			type: el.getAttribute('type') || '',
			href: el.getAttribute('href') || '',
		})),
	};
})()`
