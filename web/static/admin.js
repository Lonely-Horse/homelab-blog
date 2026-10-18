/* =========================================================
   LonelyHorse — 后台脚本（§13.3）
   职责 1/4：删除（fetch DELETE，契约见技术文档 §7.4）
   职责 2/4：实时预览（POST /admin/preview，后端待实现）
   职责 3/4：编辑页快捷键与未保存提醒
   职责 4/4：图片上传  —— TODO：POST /admin/upload
   CSP：script-src 'self'，不得内联事件
   ========================================================= */
(function () {
  "use strict";

  function bindDelete() {
    var buttons = document.querySelectorAll(".js-delete");
    Array.prototype.forEach.call(buttons, function (btn) {
      btn.addEventListener("click", function () {
        var url = btn.dataset.url;
        if (!url) return;
        if (!window.confirm("确定删除这条记录？")) return;

        btn.disabled = true;

        fetch(url, { method: "DELETE" })
          .then(function (res) {
            switch (res.status) {
              case 204: // 删除成功
                window.location.reload();
                break;
              case 400: // id 非法
                window.alert("参数错误");
                btn.disabled = false;
                break;
              case 401: // 未登录 / 会话过期
                window.location.href = "/admin/login";
                break;
              case 404: // 记录不存在（可能已被删）
                window.location.reload();
                break;
              default: // 500 及其他
                window.alert("删除失败，请重试");
                btn.disabled = false;
            }
          })
          .catch(function () {
            window.alert("网络异常，请重试");
            btn.disabled = false;
          });
      });
    });
  }

  /* ---------- 编辑页：实时预览 + 快捷键 + 未保存提醒 ---------- */
  function bindEditor() {
    var form = document.getElementById("post-form");
    if (!form) return;

    var content = form.querySelector('textarea[name="content_md"]');
    var preview = document.getElementById("preview-body");
    var previewStatus = document.getElementById("preview-status");
    if (!content || !preview || !previewStatus) return;

    var dirty = false;
    var timer = null;
    var inFlight = null;
    var lastRendered = null;

    form.addEventListener("input", function () {
      dirty = true;
    });
    form.addEventListener("submit", function () {
      dirty = false;
    });

    // 有未保存修改时，关闭/刷新页面前由浏览器弹确认框
    window.addEventListener("beforeunload", function (e) {
      if (!dirty) return;
      e.preventDefault();
      e.returnValue = "";
    });

    // Ctrl/Cmd + S：沿用当前状态保存，避免把已发布文章误降级成草稿
    document.addEventListener("keydown", function (e) {
      if (!(e.ctrlKey || e.metaKey) || String(e.key).toLowerCase() !== "s") return;
      e.preventDefault();

      var want = form.dataset.status === "published" ? "published" : "draft";
      var btn = form.querySelector(
        'button[type="submit"][name="status"][value="' + want + '"]'
      );
      if (btn) btn.click();
    });

    function renderPreview() {
      var md = content.value;
      if (md === lastRendered) return;

      if (md.trim() === "") {
        preview.innerHTML = '<p class="muted">正文渲染结果将显示在这里。</p>';
        previewStatus.textContent = "";
        lastRendered = md;
        return;
      }

      if (inFlight) inFlight.abort();
      inFlight = new AbortController();

      previewStatus.textContent = "渲染中…";

      fetch("/admin/preview", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8" },
        body: "content_md=" + encodeURIComponent(md),
        signal: inFlight.signal
      })
        .then(function (res) {
          if (res.status === 401) {
            window.location.href = "/admin/login";
            return null;
          }
          if (res.status === 404) throw new Error("preview_not_ready");
          if (!res.ok) throw new Error("preview_failed");
          return res.json();
        })
        .then(function (data) {
          if (!data) return;
          // 契约：后端返回已转义、已过滤危险 URL 的 HTML，前端直接替换
          preview.innerHTML = data.html || "";
          lastRendered = md;
          previewStatus.textContent = "";
        })
        .catch(function (err) {
          if (err && err.name === "AbortError") return;
          previewStatus.textContent =
            err && err.message === "preview_not_ready" ? "预览接口未接入" : "预览失败";
        });
    }

    // 输入停顿 450ms 再请求，避免每敲一个字都打一次后端
    content.addEventListener("input", function () {
      if (timer) clearTimeout(timer);
      timer = setTimeout(renderPreview, 450);
    });

    renderPreview();
  }

  document.addEventListener("DOMContentLoaded", function () {
    bindDelete();
    bindEditor();
    // TODO: bindUpload();
  });
})();
