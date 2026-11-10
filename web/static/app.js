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

  /* ---------- 启动 ---------- */
  // 主题尽量早应用，减少首屏闪烁（CSP 禁内联脚本，无法在 <head> 同步执行）
  var saved = readTheme();
  if (saved) applyTheme(saved);

  document.addEventListener("DOMContentLoaded", function () {
    bindThemeToggle();
    bindBack();
    bindCurtain();
  });
})();
