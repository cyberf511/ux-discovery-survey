/* لوحة الإدارة: الأسئلة، النتائج، التصدير، وقفل الاستبيان. */
(function () {
  "use strict";

  var app = document.getElementById("app");
  var st = { questions: [], sections: [], stats: [], open: true, categories: [] };

  function el(tag, cls, text) {
    var n = document.createElement(tag);
    if (cls) n.className = cls;
    if (text != null) n.textContent = text;
    return n;
  }

  function clone(id) { return document.getElementById(id).content.cloneNode(true); }

  function render(node) { app.textContent = ""; app.appendChild(node); }

  function api(method, path, body) {
    return fetch(path, {
      method: method,
      headers: body ? { "Content-Type": "application/json" } : {},
      body: body ? JSON.stringify(body) : undefined
    }).then(function (res) {
      if (!res.ok) {
        return res.json().catch(function () { return {}; }).then(function (j) {
          var e = new Error(j.error || "خطأ في الخادم");
          e.status = res.status;
          throw e;
        });
      }
      return res.json();
    });
  }

  var KINDS = {
    long_text: "نص طويل",
    short_text: "نص قصير",
    boolean: "صح / خطأ",
    single_choice: "اختيار واحد",
    multi_choice: "اختيار متعدد",
    ranking: "ترتيب قائمة",
    scale_1_5: "مقياس ١–٥"
  };

  function catLabel(v) {
    for (var i = 0; i < st.categories.length; i++) {
      if (st.categories[i].value === v) return st.categories[i].label;
    }
    return v;
  }

  // ---------- تسجيل الدخول ----------

  function showLogin(msg) {
    var node = clone("tpl-login");
    var pass = node.querySelector("[data-pass]");
    var btn = node.querySelector("[data-login]");
    var status = node.querySelector("[data-status]");

    function submit() {
      btn.disabled = true;
      api("POST", "/api/admin/login", { password: pass.value })
        .then(load)
        .catch(function (err) {
          btn.disabled = false;
          status.className = "status err";
          status.textContent = err.message;
        });
    }
    btn.addEventListener("click", submit);
    pass.addEventListener("keydown", function (e) { if (e.key === "Enter") submit(); });

    render(node);
    if (msg) { status.className = "status err"; status.textContent = msg; }
    pass.focus();
  }

  // ---------- اللوحة ----------

  function load() {
    return api("GET", "/api/admin/state").then(function (data) {
      st.questions = data.questions;
      st.sections = data.sections || [];
      st.stats = data.stats;
      st.open = data.open;
      st.categories = data.categories;
      showPanel();
    }).catch(function (err) {
      if (err.status === 401) showLogin();
      else showLogin(err.message);
    });
  }

  var activeTab = "results";

  function showPanel() {
    var node = clone("tpl-panel");

    node.querySelector("[data-openstate]").textContent = st.open
      ? "الاستبيان مفتوح ويستقبل إجابات."
      : "الاستبيان مقفل — لا تُقبل إجابات جديدة.";

    node.querySelector("[data-logout]").addEventListener("click", function () {
      api("POST", "/api/admin/logout", {}).then(function () { showLogin(); });
    });

    // الإحصاءات
    var stats = node.querySelector("[data-stats]");
    var totalStarted = 0, totalDone = 0;
    st.stats.forEach(function (s) { totalStarted += s.started; totalDone += s.finished; });
    stats.appendChild(statCard("الإجمالي", totalStarted, totalDone + " مكتملة"));
    st.stats.forEach(function (s) {
      stats.appendChild(statCard(s.label, s.started, s.finished + " مكتملة"));
    });

    // التبويبات
    var tabs = node.querySelectorAll(".tab");
    var panels = node.querySelectorAll(".panel");
    tabs.forEach(function (t) {
      t.setAttribute("aria-selected", t.dataset.tab === activeTab ? "true" : "false");
      t.addEventListener("click", function () {
        activeTab = t.dataset.tab;
        tabs.forEach(function (x) { x.setAttribute("aria-selected", x === t ? "true" : "false"); });
        panels.forEach(function (p) {
          if (p.dataset.panel === activeTab) p.setAttribute("data-active", "");
          else p.removeAttribute("data-active");
        });
        if (activeTab === "results") loadResults();
        if (activeTab === "sessions") loadSessions();
      });
    });
    panels.forEach(function (p) {
      if (p.dataset.panel === activeTab) p.setAttribute("data-active", "");
      else p.removeAttribute("data-active");
    });

    buildSectionsTab(node);
    buildQuestionsTab(node);
    buildExportTab(node);
    buildSettingsTab(node);

    // مُصفّي النتائج
    var filter = node.querySelector("[data-filter]");
    fillCategorySelect(filter, "كل الفئات");
    filter.addEventListener("change", function () { loadResults(); });
    node.querySelector("[data-refresh]").addEventListener("click", function () { load(); });

    render(node);
    if (activeTab === "results") loadResults();
    if (activeTab === "sessions") loadSessions();
  }

  function statCard(label, num, sub) {
    var c = el("div", "stat");
    c.appendChild(el("div", "label", label));
    c.appendChild(el("div", "num", String(num)));
    c.appendChild(el("div", "sub2", sub));
    return c;
  }

  function fillCategorySelect(sel, allLabel) {
    sel.textContent = "";
    if (allLabel) {
      var o = el("option", null, allLabel);
      o.value = "";
      sel.appendChild(o);
    }
    st.categories.forEach(function (c) {
      var o = el("option", null, c.label);
      o.value = c.value;
      sel.appendChild(o);
    });
  }

  // ---------- تبويب النتائج ----------

  // تبحث دائمًا داخل app لا داخل الـfragment المُمرَّر لـrender،
  // لأن render يفرّغ الـfragment فتصبح استعلاماته لاحقًا فارغة بلا خطأ.
  function loadResults() {
    var filter = app.querySelector("[data-filter]");
    var box = app.querySelector("[data-results]");
    if (!box) return;
    box.textContent = "";
    box.appendChild(el("p", "small", "جارٍ التحميل…"));
    var cat = filter ? filter.value : "";
    api("GET", "/api/admin/results" + (cat ? "?category=" + encodeURIComponent(cat) : ""))
      .then(function (data) {
        box.textContent = "";
        if (!data.rows.length) {
          box.appendChild(el("p", "small", "لا توجد أسئلة."));
          return;
        }
        data.rows.forEach(function (row) {
          var card = el("div", "card");
          var h = el("h3", null, row.question.text);
          h.style.marginTop = "0";
          card.appendChild(h);
          var meta = el("div", "small");
          if (row.question.section) meta.appendChild(el("span", "chip", row.question.section));
          meta.appendChild(el("span", "chip", KINDS[row.question.kind] || row.question.kind));
          meta.appendChild(el("span", null, row.answers.length + " إجابة"));
          card.appendChild(meta);

          if (row.answers.length) {
            var list = el("div", "answers-list");
            row.answers.forEach(function (a) {
              var d = el("div", "ans");
              d.appendChild(el("div", "who", (a.name || "بدون اسم") + " — " + a.category));
              d.appendChild(el("div", null, a.text));
              list.appendChild(d);
            });
            card.appendChild(list);
          }
          box.appendChild(card);
        });
      })
      .catch(function (err) {
        box.textContent = "";
        box.appendChild(el("p", "warn", err.message));
      });
  }

  // ---------- تبويب المشاركين ----------

  function loadSessions() {
    var box = app.querySelector("[data-sessions]");
    if (!box) return;
    box.textContent = "جارٍ التحميل…";
    api("GET", "/api/admin/results").then(function (data) {
      box.textContent = "";
      if (!data.sessions.length) {
        box.appendChild(el("p", "small", "لا يوجد مشاركون بعد."));
        return;
      }
      var table = el("table");
      var thead = el("tr");
      ["الاسم", "الفئة", "البداية", "الإجابات", "مكتملة؟", ""].forEach(function (t) {
        thead.appendChild(el("th", null, t));
      });
      table.appendChild(thead);
      data.sessions.forEach(function (s) {
        var tr = el("tr");
        tr.appendChild(el("td", null, s.name || "بدون اسم"));
        tr.appendChild(el("td", null, s.label));
        tr.appendChild(el("td", null, fmtTime(s.started_at)));
        tr.appendChild(el("td", null, String(s.answer_count)));
        tr.appendChild(el("td", null, s.done ? "نعم" : "لا"));
        var td = el("td");
        var del = el("button", null, "حذف");
        del.className = "btn ghost";
        del.style.minHeight = "36px";
        del.addEventListener("click", function () {
          if (!confirm("حذف هذا المشارك وكل إجاباته؟ لا يمكن التراجع.")) return;
          api("DELETE", "/api/admin/sessions/" + s.id).then(function () { load(); });
        });
        td.appendChild(del);
        tr.appendChild(td);
        table.appendChild(tr);
      });
      box.appendChild(table);
    }).catch(function (err) { box.textContent = err.message; });
  }

  function fmtTime(iso) {
    if (!iso) return "—";
    var d = new Date(iso);
    if (isNaN(d.getTime())) return iso;
    return d.toLocaleString("ar", { dateStyle: "short", timeStyle: "short" });
  }

  // ---------- تبويب الأقسام ----------

  function buildSectionsTab(root) {
    var list = root.querySelector("[data-sections-list]");
    var summary = root.querySelector("[data-sec-summary]");
    if (!list) return;

    var activeQ = 0, totalQ = 0, activeSecs = 0;
    st.sections.forEach(function (s) {
      activeQ += s.active;
      totalQ += s.total;
      if (s.active > 0) activeSecs++;
    });
    summary.textContent = activeSecs + " من " + st.sections.length + " أقسام مفعّلة — " +
      activeQ + " سؤالًا من " + totalQ + " تظهر للمشاركين";

    var table = el("table");
    var head = el("tr");
    ["القسم", "الأسئلة", "الحالة", ""].forEach(function (t) { head.appendChild(el("th", null, t)); });
    table.appendChild(head);

    st.sections.forEach(function (sec) {
      var on = sec.active > 0;
      var tr = el("tr");
      tr.appendChild(el("td", null, sec.name));
      tr.appendChild(el("td", null, String(sec.total)));
      var status = el("td");
      status.appendChild(el("span", "chip", on ? "مفعّل" : "معطّل"));
      tr.appendChild(status);
      var act = el("td");
      var b = el("button", "btn ghost", on ? "تعطيل" : "تشغيل");
      b.style.minHeight = "36px";
      b.addEventListener("click", function () { toggleSection(sec.name, !on); });
      act.appendChild(b);
      tr.appendChild(act);
      if (!on) tr.style.opacity = ".55";
      table.appendChild(tr);
    });

    list.textContent = "";
    list.appendChild(table);

    root.querySelector("[data-sec-none]").addEventListener("click", function () {
      bulkSections(false);
    });
    root.querySelector("[data-sec-all]").addEventListener("click", function () {
      bulkSections(true);
    });
  }

  function toggleSection(name, active) {
    activeTab = "sections";
    api("POST", "/api/admin/sections/toggle", { section: name, active: active })
      .then(load)
      .catch(function (e) { alert(e.message); });
  }

  function bulkSections(active) {
    var what = active ? "تشغيل" : "تعطيل";
    if (!confirm(what + " كل الأقسام؟")) return;
    activeTab = "sections";
    var chain = Promise.resolve();
    st.sections.forEach(function (sec) {
      chain = chain.then(function () {
        return api("POST", "/api/admin/sections/toggle", { section: sec.name, active: active });
      });
    });
    chain.then(load).catch(function (e) { alert(e.message); });
  }

  // ---------- تبويب الأسئلة ----------

  function catCheckboxes(container) {
    container.textContent = "";
    var boxes = [];
    st.categories.forEach(function (c) {
      var lbl = el("label", "row");
      lbl.style.gap = "4px";
      var cb = el("input");
      cb.type = "checkbox";
      cb.value = c.value;
      lbl.appendChild(cb);
      lbl.appendChild(el("span", "small", c.label));
      container.appendChild(lbl);
      boxes.push(cb);
    });
    return function () {
      return boxes.filter(function (b) { return b.checked; }).map(function (b) { return b.value; });
    };
  }

  function buildQuestionsTab(root) {
    var text = root.querySelector("[data-qtext]");
    var section = root.querySelector("[data-qsection]");
    var kind = root.querySelector("[data-qkind]");
    var req = root.querySelector("[data-qreq]");
    var opts = root.querySelector("[data-qopts]");
    var status = root.querySelector("[data-qstatus]");
    var getCats = catCheckboxes(root.querySelector("[data-qcats]"));

    root.querySelector("[data-qadd]").addEventListener("click", function () {
      api("POST", "/api/admin/questions", {
        text: text.value,
        kind: kind.value,
        section: section.value,
        required: req.checked,
        categories: getCats(),
        options: splitOptions(opts.value)
      }).then(function () {
        text.value = ""; opts.value = "";
        load();
      }).catch(function (err) {
        status.className = "status err";
        status.textContent = err.message;
      });
    });

    var getICats = catCheckboxes(root.querySelector("[data-icats]"));
    var importBox = root.querySelector("[data-import]");
    root.querySelector("[data-importbtn]").addEventListener("click", function () {
      if (!importBox.value.trim()) return;
      api("POST", "/api/admin/questions/import", {
        text: importBox.value,
        categories: getICats()
      }).then(function (r) {
        importBox.value = "";
        alert("أُضيف " + r.added + " سؤالًا.");
        load();
      }).catch(function (err) { alert(err.message); });
    });

    var list = root.querySelector("[data-qlist]");
    st.questions.forEach(function (q, i) {
      list.appendChild(questionRow(q, i));
    });
  }

  function splitOptions(s) {
    return s.split(/[,،\n]/).map(function (x) { return x.trim(); }).filter(Boolean);
  }

  function questionRow(q, i) {
    var row = el("div", "qrow");
    row.appendChild(el("span", "pos", String(i + 1)));

    var body = el("div", "body");
    body.appendChild(el("div", null, q.text + (q.required ? " *" : "")));
    var meta = el("div", "meta");
    if (q.section) meta.appendChild(el("span", "chip", q.section));
    meta.appendChild(el("span", "chip", KINDS[q.kind] || q.kind));
    if (q.categories && q.categories.length) {
      meta.appendChild(el("span", null, q.categories.map(catLabel).join("، ")));
    } else {
      meta.appendChild(el("span", null, "كل الفئات"));
    }
    if (q.options && q.options.length) {
      meta.appendChild(el("div", null, "الخيارات: " + q.options.join(" | ")));
    }
    body.appendChild(meta);
    row.appendChild(body);

    var tools = el("div", "tools");
    tools.appendChild(toolBtn("▲", function () { moveQuestion(i, -1); }, i === 0));
    tools.appendChild(toolBtn("▼", function () { moveQuestion(i, 1); }, i === st.questions.length - 1));
    tools.appendChild(toolBtn("تعديل النص", function () {
      var t = prompt("نص السؤال:", q.text);
      if (t == null || !t.trim()) return;
      api("PUT", "/api/admin/questions/" + q.id, {
        text: t, kind: q.kind, section: q.section,
        required: q.required, categories: q.categories, options: q.options
      }).then(load).catch(function (e) { alert(e.message); });
    }));
    tools.appendChild(toolBtn("حذف", function () {
      if (!confirm("حذف السؤال؟ الإجابات المجموعة عليه تبقى في التصدير.")) return;
      api("DELETE", "/api/admin/questions/" + q.id).then(load).catch(function (e) { alert(e.message); });
    }));
    row.appendChild(tools);
    return row;
  }

  function toolBtn(label, fn, disabled) {
    var b = el("button", null, label);
    b.type = "button";
    b.disabled = !!disabled;
    b.addEventListener("click", fn);
    return b;
  }

  function moveQuestion(i, delta) {
    var j = i + delta;
    if (j < 0 || j >= st.questions.length) return;
    var ids = st.questions.map(function (q) { return q.id; });
    var t = ids[i]; ids[i] = ids[j]; ids[j] = t;
    activeTab = "questions";
    api("POST", "/api/admin/questions/reorder", { ids: ids }).then(load);
  }

  // ---------- تبويب التصدير ----------

  function buildExportTab(root) {
    var sel = root.querySelector("[data-expcat]");
    fillCategorySelect(sel, null);
    var link = root.querySelector("[data-exportcat]");
    function sync() { link.href = "/api/admin/export.csv?category=" + encodeURIComponent(sel.value); }
    sel.addEventListener("change", sync);
    sync();
  }

  // ---------- تبويب الإعدادات ----------

  function buildSettingsTab(root) {
    var btn = root.querySelector("[data-toggleopen]");
    btn.textContent = st.open ? "قفل الاستبيان" : "إعادة فتح الاستبيان";
    btn.addEventListener("click", function () {
      var next = !st.open;
      if (next === false && !confirm("قفل الاستبيان؟ لن تُقبل إجابات جديدة.")) return;
      activeTab = "settings";
      api("POST", "/api/admin/open", { open: next }).then(load).catch(function (e) { alert(e.message); });
    });
    root.querySelector("[data-link]").value = location.origin + "/";
  }

  load();
})();
