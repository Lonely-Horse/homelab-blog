/* =========================================================
   LonelyHorse — 全局脚本
   职责：主题切换（记忆于 localStorage）+ 首页幕布交互
   CSP：script-src 'self'，禁止内联事件，全部在外部文件里 addEventListener
   ========================================================= */
(function () {
  "use strict";

  /* ---------- 主题 ---------- */
  var THEME_KEY = "hb-theme";

  function applyTheme(theme) {
    document.documentElement.dataset.theme = theme === "light" ? "light" : "dark";
  }

  function readTheme() {
    try {
      return localStorage.getItem(THEME_KEY);
    } catch (e) {
      return null;
    }
  }

  function saveTheme(theme) {
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch (e) {
      /* 隐私模式下 localStorage 可能不可用，忽略即可 */
    }
  }

  function currentTheme() {
    return document.documentElement.dataset.theme === "light" ? "light" : "dark";
  }

  function bindThemeToggle() {
    var buttons = document.querySelectorAll("[data-theme-toggle]");
    Array.prototype.forEach.call(buttons, function (btn) {
      btn.addEventListener("click", function () {
        var next = currentTheme() === "light" ? "dark" : "light";
        applyTheme(next);
        saveTheme(next);
        btn.setAttribute("aria-label", next === "light" ? "切换到暗色主题" : "切换到亮色主题");
      });
    });
  }

  /* ---------- 浮动返回按钮 ---------- */
  function bindBack() {
    var btns = document.querySelectorAll("[data-back]");
    Array.prototype.forEach.call(btns, function (btn) {
      btn.addEventListener("click", function () {
        var fallback = btn.getAttribute("data-fallback") || "/posts";
        // history.length 含当前页；>1 才认为有可返回的上一条，否则退到 fallback
        if (window.history.length > 1) {
          window.history.back();
        } else {
          window.location.href = fallback;
        }
      });
    });
  }

  /* ---------- 首页幕布：点击 / 键盘 / 上拉 拉起 ---------- */
  function bindCurtain() {
    var curtain = document.getElementById("curtain");
    if (!curtain) return;

    var enterUrl = curtain.dataset.enter || "/posts";
    var opened = false;

    function open() {
      if (opened) return;
      opened = true;
      curtain.classList.remove("is-dragging");
      // 清掉手势期间写入的行内 transform：行内样式优先级高于 .is-open，
      // 残留一次平板上的“异常中断”就会把幕布顶偏、露出底色
      curtain.style.transform = "";
      curtain.style.setProperty("--drag", "0");
      curtain.classList.add("is-open");

      // 动画结束后进入文章列表；transitionend + 定时器双保险
      var timer = null;
      function go() {
        if (timer) clearTimeout(timer);
        curtain.removeEventListener("transitionend", go);
        window.location.href = enterUrl;
      }
      curtain.addEventListener("transitionend", go);
      timer = setTimeout(go, 1600);
    }

    // 点击 / 回车 / 空格
    curtain.addEventListener("click", open);
    curtain.addEventListener("keydown", function (e) {
      if (e.key === "Enter" || e.key === " " || e.key === "Spacebar") {
        e.preventDefault();
        open();
      }
    });

    // 上拉手势
    var startY = 0;
    var dy = 0;
    var dragging = false;

    // pointermove 可能每帧触发多次，用 rAF 合并成每帧一次写入，避免抖动
    var rafId = null;
    function paint() {
      rafId = null;
      // 幕布本身上移；同时把 0~1 的进度写进 --drag，
      // 让 .curtain-inner 的 opacity/transform 跟着做视差淡出（CSS 侧计算）
      curtain.style.transform = "translate3d(0," + dy + "px,0)";
      var progress = Math.min(1, Math.abs(dy) / (window.innerHeight * 0.5));
      curtain.style.setProperty("--drag", String(progress));
    }

    curtain.addEventListener("pointerdown", function (e) {
      if (opened) return;
      dragging = true;
      startY = e.clientY;
      dy = 0;
      curtain.classList.add("is-dragging");
      if (curtain.setPointerCapture) {
        try { curtain.setPointerCapture(e.pointerId); } catch (err) {}
      }
    });

    curtain.addEventListener("pointermove", function (e) {
      if (!dragging || opened) return;
      dy = e.clientY - startY;
      if (dy > 0) dy = 0; // 只允许向上拉
      if (rafId === null) rafId = window.requestAnimationFrame(paint);
    });

    function endDrag() {
      if (!dragging) return;
      dragging = false;
      if (rafId !== null) {
        window.cancelAnimationFrame(rafId);
        rafId = null;
      }

      var threshold = window.innerHeight * 0.18;
      if (-dy > threshold) {
        // 先解除拖拽态的瞬时定位，再交回 CSS 过渡到 -100%
        curtain.classList.remove("is-dragging");
        window.requestAnimationFrame(function () {
          curtain.style.transform = "";
          curtain.style.setProperty("--drag", "0");
          open();
        });
      } else {
        // 未过阈值：回弹（--drag 归零，内容随过渡淡回）
        curtain.classList.remove("is-dragging");
        curtain.style.transform = "";
        curtain.style.setProperty("--drag", "0");
      }
    }

    curtain.addEventListener("pointerup", endDrag);
    curtain.addEventListener("pointercancel", endDrag);
    // 触摸被系统/浏览器打断时可能只发 lostpointercapture：
    // 这里只做归位（不判定是否进入），避免残留位移把幕布顶偏
    curtain.addEventListener("lostpointercapture", function () {
      if (!dragging) return;
      dragging = false;
      if (rafId !== null) {
        window.cancelAnimationFrame(rafId);
        rafId = null;
      }
      curtain.classList.remove("is-dragging");
      curtain.style.transform = "";
      curtain.style.setProperty("--drag", "0");
    });
  }

  /* ---------- 文章目录：折叠子树 + 滚动高亮（§20.6） ----------
     结构与 id 全在后端 toc_html 里；这里只加按钮与状态类。
     不用 <details>/<summary>（链接与折叠不可分离），不用 IntersectionObserver
     （答的是"是否可见"，不是"读到哪节"）。 */
  function bindTOC() {
    var nav = document.querySelector(".toc .toc-nav");
    var body = document.querySelector(".article-body");
    if (!nav || !body) return;

    // 折叠：有子列表的 li 前置一个 <button>，点按钮只折叠、点链接只跳转
    var items = nav.querySelectorAll("li");
    Array.prototype.forEach.call(items, function (li) {
      var sub = null;
      for (var i = 0; i < li.children.length; i++) {
        if (li.children[i].tagName === "OL") { sub = li.children[i]; break; }
      }
      if (!sub) return;
      li.classList.add("has-children");
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "toc-toggle";
      btn.setAttribute("aria-expanded", "true");
      btn.setAttribute("aria-label", "折叠或展开该小节");
      btn.addEventListener("click", function () {
        var collapsed = li.classList.toggle("is-collapsed");
        btn.setAttribute("aria-expanded", collapsed ? "false" : "true");
      });
      li.insertBefore(btn, li.firstChild);
    });

    var links = nav.querySelectorAll('a[href^="#"]');
    var headings = body.querySelectorAll("h2[id], h3[id], h4[id]");
    if (!links.length || !headings.length) return;

    // offsetTop 是相对最近定位祖先的（.article 是 relative），
    // 必须沿 offsetParent 链累加成文档坐标，才能与 scrollY 比较
    function docTop(el) {
      var t = 0;
      while (el) { t += el.offsetTop; el = el.offsetParent; }
      return t;
    }
    var tops = [];
    function measure() {
      tops = [];
      Array.prototype.forEach.call(headings, function (h) { tops.push(docTop(h)); });
    }
    measure();

    var OFFSET = 96; // 与 scroll-margin-top 对齐
    var ticking = false;

    function update() {
      ticking = false;
      var y = window.scrollY + OFFSET;
      var current = null;
      for (var i = 0; i < headings.length; i++) {
        if (tops[i] <= y) { current = headings[i].id; } else { break; }
      }
      // 页底最后一节可能永远够不到 OFFSET 线：触底时强制激活最后一条
      var doc = document.documentElement;
      if (window.scrollY + window.innerHeight >= doc.scrollHeight - 2) {
        current = headings[headings.length - 1].id;
      }
      Array.prototype.forEach.call(links, function (a) {
        if (current && a.getAttribute("href") === "#" + current) {
          a.setAttribute("aria-current", "true");
        } else {
          a.removeAttribute("aria-current");
        }
      });
    }

    function onScroll() {
      if (!ticking) { ticking = true; window.requestAnimationFrame(update); }
    }
    window.addEventListener("scroll", onScroll, { passive: true });
    // resize 后布局坐标会变：必须重测 tops 再刷新高亮，否则锚点错位
    window.addEventListener("resize", function () {
      measure();
      onScroll();
    }, { passive: true });
    update();
  }

  /* ---------- 启动 ---------- */
  // 主题尽量早应用，减少首屏闪烁（CSP 禁内联脚本，无法在 <head> 同步执行）
  var saved = readTheme();
  if (saved) applyTheme(saved);

  document.addEventListener("DOMContentLoaded", function () {
    bindThemeToggle();
    bindBack();
    bindCurtain();
    bindTOC();
  });
})();
