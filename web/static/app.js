/* واجهة المشارك: سؤال واحد بالشاشة، حفظ تلقائي، ومقاومة لانقطاع الشبكة. */
(function () {
  "use strict";

  var KEY_SESSION = "survey.session";
  var KEY_PENDING = "survey.pending";
  var KEY_INDEX = "survey.index";

  var app = document.getElementById("app");
  var state = {
    sessionId: null,
    questions: [],
    answers: {},   // questionId -> value
    index: 0,
    open: true,
    categories: []
  };
  var dirty = false; // هل قيمة السؤال الحالي تغيّرت ولم تُحفظ بعد

  // ---------- أدوات ----------

  function el(tag, cls, text) {
    var n = document.createElement(tag);
    if (cls) n.className = cls;
    if (text != null) n.textContent = text;
    return n;
  }

  function clone(id) {
    return document.getElementById(id).content.cloneNode(true);
  }

  function render(node) {
    app.textContent = "";
    app.appendChild(node);
  }

  function api(method, path, body) {
    return fetch(path, {
      method: method,
      headers: body ? { "Content-Type": "application/json" } : {},
      body: body ? JSON.stringify(body) : undefined
    }).then(function (res) {
      if (!res.ok) {
        return res.json().catch(function () { return {}; }).then(function (j) {
          var e = new Error(j.error || "تعذّر الاتصال بالخادم");
          e.status = res.status;
          throw e;
        });
      }
      return res.status === 204 ? null : res.json();
    });
  }

  function store(key, val) {
    try {
      if (val === undefined) return localStorage.getItem(key);
      if (val === null) localStorage.removeItem(key);
      else localStorage.setItem(key, val);
    } catch (e) { /* التخزين المحلي قد يكون معطّلًا؛ التطبيق يعمل بدونه */ }
    return null;
  }

  // ---------- طابور الحفظ ----------
  // كل إجابة تُحفظ محليًا أولًا ثم تُرسل. الفشل يبقيها في الطابور لإعادة المحاولة.

  function pending() {
    try { return JSON.parse(store(KEY_PENDING) || "{}"); } catch (e) { return {}; }
  }

  function queueAnswer(qid, value) {
    var p = pending();
    p[qid] = value;
    store(KEY_PENDING, JSON.stringify(p));
  }

  function dequeue(qid) {
    var p = pending();
    delete p[qid];
    store(KEY_PENDING, JSON.stringify(p));
  }

  function flush(statusEl) {
    var p = pending();
    var ids = Object.keys(p);
    if (!ids.length) return Promise.resolve(true);
    if (statusEl) setStatus(statusEl, "جارٍ الحفظ…", "");
    var chain = Promise.resolve(true);
    ids.forEach(function (qid) {
      chain = chain.then(function (okSoFar) {
        return api("POST", "/api/sessions/" + state.sessionId + "/answers", {
          question_id: Number(qid),
          value: p[qid]
        }).then(function () {
          dequeue(qid);
          return okSoFar;
        }).catch(function (err) {
          if (err.status === 404 || err.status === 409) {
            // الجلسة غير موجودة أو الاستبيان مقفل: لا فائدة من إعادة المحاولة.
            dequeue(qid);
          }
          if (statusEl) setStatus(statusEl, err.message + " — سنعيد المحاولة تلقائيًا", "err");
          return false;
        });
      });
    });
    return chain.then(function (ok) {
      if (ok && statusEl) setStatus(statusEl, "تم الحفظ ✓", "ok");
      return ok;
    });
  }

  function setStatus(node, text, cls) {
    if (!node) return;
    node.textContent = text;
    node.className = "status " + (cls || "");
  }

  // إعادة محاولة دورية للإجابات العالقة عند عودة الشبكة.
  setInterval(function () {
    if (state.sessionId && Object.keys(pending()).length) flush(document.querySelector("[data-status]"));
  }, 15000);
  window.addEventListener("online", function () {
    if (state.sessionId) flush(document.querySelector("[data-status]"));
  });

  // ---------- شاشة البداية ----------

  function showStart() {
    var node = clone("tpl-start");
    var catsBox = node.querySelector("[data-cats]");
    var nameInput = node.querySelector("[data-name]");
    var beginBtn = node.querySelector("[data-begin]");
    var chosen = null;

    state.categories.forEach(function (c) {
      var b = el("button", "cat", c.label);
      b.type = "button";
      b.setAttribute("aria-pressed", "false");
      b.addEventListener("click", function () {
        chosen = c.value;
        catsBox.querySelectorAll(".cat").forEach(function (x) {
          x.setAttribute("aria-pressed", "false");
        });
        b.setAttribute("aria-pressed", "true");
        beginBtn.disabled = false;
      });
      catsBox.appendChild(b);
    });

    beginBtn.addEventListener("click", function () {
      if (!chosen) return;
      beginBtn.disabled = true;
      beginBtn.textContent = "جارٍ البدء…";
      api("POST", "/api/sessions", { category: chosen, name: nameInput.value.trim() })
        .then(function (sess) {
          store(KEY_SESSION, sess.id);
          store(KEY_INDEX, "0");
          return loadSession(sess.id);
        })
        .catch(function (err) {
          beginBtn.disabled = false;
          beginBtn.textContent = "ابدأ";
          alert(err.message);
        });
    });

    render(node);

    if (!state.open) {
      var warn = el("div", "closed", "الاستبيان مقفل حاليًا ولا يستقبل إجابات جديدة.");
      app.insertBefore(warn, app.firstChild);
      beginBtn.disabled = true;
    }
  }

  // ---------- شاشة السؤال ----------

  function currentQuestion() { return state.questions[state.index]; }

  function showQuestion() {
    var q = currentQuestion();
    if (!q) { showDone(); return; }

    var node = clone("tpl-question");
    var bar = node.querySelector("[data-bar]");
    var counter = node.querySelector("[data-counter]");
    var section = node.querySelector("[data-section]");
    var qtext = node.querySelector("[data-qtext]");
    var answerBox = node.querySelector("[data-answer]");
    var prev = node.querySelector("[data-prev]");
    var next = node.querySelector("[data-next]");
    var skip = node.querySelector("[data-skip]");
    var status = node.querySelector("[data-status]");

    bar.style.width = Math.round((state.index / state.questions.length) * 100) + "%";
    counter.textContent = (state.index + 1) + " من " + state.questions.length;
    section.textContent = q.section || "";
    qtext.textContent = q.text + (q.required ? " *" : "");

    buildAnswer(answerBox, q, function (val) {
      state.answers[q.id] = val;
      dirty = true;
      next.disabled = q.required && isEmpty(val);
    });

    prev.disabled = state.index === 0;
    next.disabled = q.required && isEmpty(state.answers[q.id]);
    skip.style.display = q.required ? "none" : "";
    next.textContent = state.index === state.questions.length - 1 ? "إنهاء" : "التالي";

    prev.addEventListener("click", function () { saveThen(status, -1); });
    next.addEventListener("click", function () { saveThen(status, 1); });
    skip.addEventListener("click", function () { move(1); });

    render(node);
    if (Object.keys(pending()).length) flush(status);
  }

  function isEmpty(v) {
    if (v === undefined || v === null || v === "") return true;
    if (Array.isArray(v) && v.length === 0) return true;
    return false;
  }

  function saveThen(status, delta) {
    var q = currentQuestion();
    var val = state.answers[q.id];
    if (!isEmpty(val) && dirty) {
      queueAnswer(q.id, val);
      dirty = false;
      flush(status);
    }
    move(delta);
  }

  function move(delta) {
    var next = state.index + delta;
    if (next < 0) next = 0;
    dirty = false;
    if (next >= state.questions.length) { finish(); return; }
    state.index = next;
    store(KEY_INDEX, String(next));
    showQuestion();
  }

  function finish() {
    flush(null).then(function () {
      return api("POST", "/api/sessions/" + state.sessionId + "/finish", {});
    }).catch(function () { /* الإنهاء تجميلي؛ الإجابات محفوظة أصلًا */ })
      .then(function () {
        state.index = state.questions.length;
        store(KEY_INDEX, String(state.index));
        showDone();
      });
  }

  function showDone() {
    var node = clone("tpl-done");
    node.querySelector("[data-review]").addEventListener("click", function () {
      state.index = 0;
      store(KEY_INDEX, "0");
      showQuestion();
    });
    render(node);
  }

  // ---------- عناصر الإجابة حسب النوع ----------

  function buildAnswer(box, q, onChange) {
    box.textContent = "";
    var val = state.answers[q.id];

    if (q.kind === "long_text" || q.kind === "short_text") {
      var input = q.kind === "long_text" ? el("textarea") : el("input");
      if (q.kind === "short_text") input.type = "text";
      input.value = val == null ? "" : String(val);
      input.placeholder = "اكتب إجابتك هنا…";
      input.addEventListener("input", function () { onChange(input.value.trim()); });
      box.appendChild(input);
      return;
    }

    if (q.kind === "boolean") {
      var opts = el("div", "opts");
      [["نعم", true], ["لا", false]].forEach(function (pair) {
        var b = optButton(pair[0], val === pair[1]);
        b.addEventListener("click", function () {
          onChange(pair[1]);
          markSingle(opts, b);
        });
        opts.appendChild(b);
      });
      box.appendChild(opts);
      return;
    }

    if (q.kind === "single_choice") {
      var one = el("div", "opts");
      (q.options || []).forEach(function (o) {
        var b = optButton(o, val === o);
        b.addEventListener("click", function () {
          onChange(o);
          markSingle(one, b);
        });
        one.appendChild(b);
      });
      box.appendChild(one);
      return;
    }

    if (q.kind === "multi_choice") {
      var chosen = Array.isArray(val) ? val.slice() : [];
      box.appendChild(el("p", "hint", "تقدر تختار أكثر من إجابة."));
      var many = el("div", "opts");
      (q.options || []).forEach(function (o) {
        var b = optButton(o, chosen.indexOf(o) >= 0);
        b.addEventListener("click", function () {
          var i = chosen.indexOf(o);
          if (i >= 0) chosen.splice(i, 1); else chosen.push(o);
          b.setAttribute("aria-pressed", chosen.indexOf(o) >= 0 ? "true" : "false");
          b.querySelector(".mark").textContent = chosen.indexOf(o) >= 0 ? "✓" : "";
          onChange(chosen.slice());
        });
        many.appendChild(b);
      });
      box.appendChild(many);
      return;
    }

    if (q.kind === "scale_1_5") {
      var scale = el("div", "scale");
      [1, 2, 3, 4, 5].forEach(function (n) {
        var b = optButton(String(n), val === n);
        b.addEventListener("click", function () {
          onChange(n);
          markSingle(scale, b);
        });
        scale.appendChild(b);
      });
      box.appendChild(scale);
      return;
    }

    if (q.kind === "ranking") {
      var order = Array.isArray(val) && val.length ? val.slice() : (q.options || []).slice();
      box.appendChild(el("p", "hint", "رتّبها بالأسهم: الأعلى = الأهم."));
      var list = el("div", "rank");
      box.appendChild(list);
      drawRank(list, order, onChange);
      onChange(order.slice()); // الترتيب الابتدائي إجابة صالحة
      return;
    }

    box.appendChild(el("p", "hint", "نوع سؤال غير مدعوم: " + q.kind));
  }

  function drawRank(list, order, onChange) {
    list.textContent = "";
    order.forEach(function (label, i) {
      var row = el("div", "rank-item");
      row.appendChild(el("span", "num", String(i + 1)));
      row.appendChild(el("span", "label", label));

      var up = el("button", "move", "▲");
      up.type = "button";
      up.disabled = i === 0;
      up.title = "لأعلى";
      up.addEventListener("click", function () { swap(i, i - 1); });

      var down = el("button", "move", "▼");
      down.type = "button";
      down.disabled = i === order.length - 1;
      down.title = "لأسفل";
      down.addEventListener("click", function () { swap(i, i + 1); });

      row.appendChild(up);
      row.appendChild(down);
      list.appendChild(row);
    });

    function swap(a, b) {
      var t = order[a];
      order[a] = order[b];
      order[b] = t;
      onChange(order.slice());
      drawRank(list, order, onChange);
    }
  }

  function optButton(text, pressed) {
    var b = el("button", "opt");
    b.type = "button";
    b.setAttribute("aria-pressed", pressed ? "true" : "false");
    var mark = el("span", "mark", pressed ? "✓" : "");
    b.appendChild(mark);
    b.appendChild(el("span", null, text));
    return b;
  }

  function markSingle(container, chosen) {
    container.querySelectorAll(".opt").forEach(function (x) {
      x.setAttribute("aria-pressed", "false");
      x.querySelector(".mark").textContent = "";
    });
    chosen.setAttribute("aria-pressed", "true");
    chosen.querySelector(".mark").textContent = "✓";
  }

  // ---------- الإقلاع ----------

  function loadSession(id) {
    return api("GET", "/api/sessions/" + id).then(function (data) {
      state.sessionId = data.session.id;
      state.questions = data.questions;
      state.open = data.open;
      state.answers = {};
      Object.keys(data.answers || {}).forEach(function (k) {
        state.answers[Number(k)] = data.answers[k];
      });
      // الإجابات العالقة محليًا أحدث من نسخة الخادم.
      var p = pending();
      Object.keys(p).forEach(function (k) { state.answers[Number(k)] = p[k]; });

      var saved = parseInt(store(KEY_INDEX) || "0", 10);
      state.index = isNaN(saved) ? 0 : Math.min(Math.max(saved, 0), state.questions.length);

      if (data.session.finished_at && state.index >= state.questions.length) showDone();
      else showQuestion();
    });
  }

  api("GET", "/api/meta").then(function (meta) {
    state.open = meta.open;
    state.categories = meta.categories;
    var id = store(KEY_SESSION);
    if (!id) { showStart(); return; }
    return loadSession(id).catch(function () {
      // جلسة قديمة أو قاعدة بيانات جديدة: نبدأ من الصفر.
      store(KEY_SESSION, null);
      store(KEY_INDEX, null);
      store(KEY_PENDING, null);
      showStart();
    });
  }).catch(function (err) {
    app.textContent = "";
    app.appendChild(el("div", "closed", "تعذّر تحميل الاستبيان: " + err.message));
  });
})();
