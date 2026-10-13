/* =========================================================
   LonelyHorse — 后台脚本（§13.3）
   职责 1/3：删除（fetch DELETE，契约见技术文档 §7.4）
   职责 2/3：实时预览  —— TODO：POST /admin/preview
   职责 3/3：图片上传  —— TODO：POST /admin/upload
   CSP：script-src 'self'，不得内联事件
   ========================================================= */
(function () {
  "use strict";

  function bindDelete() {
    var buttons = document.querySelectorAll(".js-delete");
    Array.prototype.forEach.call(buttons, function (btn) {
      btn.addEventListener("click", function () {
        var id = btn.dataset.id;
        if (!window.confirm("确定删除这条记录？")) return;

        btn.disabled = true;

        fetch("/admin/posts/" + encodeURIComponent(id), { method: "DELETE" })
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

  document.addEventListener("DOMContentLoaded", function () {
    bindDelete();
    // TODO: bindPreview(); bindUpload();
  });
})();
