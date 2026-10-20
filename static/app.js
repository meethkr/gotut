/* GoTut — local interactive Go tutorial.
   No build step, no dependencies. Talks to the Go server in /api. */

(function () {
  "use strict";

  // ---------------------------------------------------------------- state

  const LS_PROGRESS = "gotut.progress.v1";
  const LS_CODES = "gotut.codes.v1";
  const LS_THEME = "gotut.theme.v1";

  const state = {
    curriculum: [],
    flat: [],          // { module, topic, step } in reading order
    index: 0,
    quiz: {},          // stepKey -> { selected:Set, checked:bool }
    hints: {},         // exerciseId -> number revealed
    solutions: {},     // exerciseId -> bool
    output: {},        // exerciseId -> { tab, data }
    progress: read(LS_PROGRESS, {}),
    codes: read(LS_CODES, {}),
  };

  function read(key, fallback) {
    try {
      const raw = localStorage.getItem(key);
      return raw ? JSON.parse(raw) : fallback;
    } catch (e) { return fallback; }
  }
  function write(key, value) {
    try { localStorage.setItem(key, JSON.stringify(value)); } catch (e) { /* ignore */ }
  }

  const $ = (sel, root) => (root || document).querySelector(sel);

  // ------------------------------------------------------------ markdown

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  // ---------------------------------------------------- syntax highlighting
  //
  // A small, dependency-free tokenizer. Each highlighter returns HTML with the
  // non-token text escaped, so the result is safe to drop into innerHTML.

  const GO_KEYWORDS = new Set([
    "break", "case", "chan", "const", "continue", "default", "defer", "else",
    "fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
    "map", "package", "range", "return", "select", "struct", "switch", "type", "var",
  ]);
  const GO_TYPES = new Set([
    "any", "bool", "byte", "complex64", "complex128", "error", "float32", "float64",
    "int", "int8", "int16", "int32", "int64", "rune", "string",
    "uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
  ]);
  const GO_BUILTINS = new Set([
    "append", "cap", "clear", "close", "complex", "copy", "delete", "imag", "len",
    "make", "max", "min", "new", "panic", "print", "println", "real", "recover",
  ]);
  const GO_CONSTANTS = new Set(["true", "false", "iota", "nil"]);

  const GO_RE = new RegExp(
    [
      "(\\/\\/[^\\n]*|\\/\\*[\\s\\S]*?\\*\\/)",                       // 1 comment
      "(`[^`]*`|\"(?:\\\\.|[^\"\\\\\\n])*\"|'(?:\\\\.|[^'\\\\\\n])*')", // 2 string / rune
      "(0[xX][0-9a-fA-F_]+|0[bB][01_]+|0[oO][0-7_]+|\\d[\\d_]*(?:\\.[\\d_]+)?(?:[eE][+-]?\\d+)?i?)", // 3 number
      "\\b([A-Za-z_]\\w*)\\b",                                        // 4 identifier
      "(\\.\\.\\.|==|!=|<=|>=|:=|<-|&&|\\|\\||\\+\\+|--|<<|>>|&\\^|[+\\-*/%&|^<>=!])", // 5 operator
    ].join("|"),
    "g"
  );

  const BASH_RE = new RegExp(
    [
      "(#[^\\n]*)",                                                    // 1 comment
      "(\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*')",             // 2 string
      "(\\$\\{[^}]*\\}|\\$\\w+|\\$\\d)",                              // 3 variable
      "\\b(go|cd|git|curl|docker|export|echo|make|set|if|then|fi|for|in|do|done|rm|mkdir|cp|mv|cat|grep|sudo|brew|apt)\\b", // 4 command
    ].join("|"),
    "g"
  );

  const SQL_KEYWORDS = new Set([
    "select", "from", "where", "insert", "into", "values", "update", "set", "delete",
    "create", "table", "if", "not", "exists", "primary", "key", "autoincrement",
    "default", "current_timestamp", "order", "by", "desc", "asc", "limit", "and",
    "or", "null", "references", "unique", "index", "on", "join", "left", "inner",
    "group", "as", "distinct", "coalesce", "sum", "count",
  ]);
  const SQL_RE = /(--[^\n]*|\/\*[\s\S]*?\*\/)|('(?:''|[^'])*')|([A-Za-z_]\w*)|(\d+)/g;

  const JSON_RE = /("(?:\\.|[^"\\])*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|\b(true|false|null)\b/g;

  // Run a global regex over src, escaping the gaps and classifying the matches.
  function tokenize(src, re, classify) {
    let out = "";
    let last = 0;
    let m;
    re.lastIndex = 0;

    while ((m = re.exec(src)) !== null) {
      if (m.index > last) out += escapeHtml(src.slice(last, m.index));
      const cls = classify(m, re.lastIndex);
      const text = escapeHtml(m[0]);
      out += cls ? '<span class="' + cls + '">' + text + "</span>" : text;
      last = m.index + m[0].length;
    }
    out += escapeHtml(src.slice(last));
    return out;
  }

  function highlightGo(src) {
    return tokenize(src, GO_RE, (m, after) => {
      if (m[1]) return "tok-com";
      if (m[2]) return "tok-str";
      if (m[3]) return "tok-num";
      if (m[4]) {
        const w = m[4];
        if (GO_KEYWORDS.has(w)) return "tok-kw";
        if (GO_TYPES.has(w)) return "tok-typ";
        if (GO_BUILTINS.has(w)) return "tok-bi";
        if (GO_CONSTANTS.has(w)) return "tok-const";
        if (src[after] === "(") return "tok-fn";
        return "";
      }
      if (m[5]) return "tok-op";
      return "";
    });
  }

  function highlightBash(src) {
    return tokenize(src, BASH_RE, (m) => {
      if (m[1]) return "tok-com";
      if (m[2]) return "tok-str";
      if (m[3]) return "tok-num";
      if (m[4]) return "tok-kw";
      return "";
    });
  }

  function highlightSQL(src) {
    return tokenize(src, SQL_RE, (m) => {
      if (m[1]) return "tok-com";
      if (m[2]) return "tok-str";
      if (m[3]) return SQL_KEYWORDS.has(m[3].toLowerCase()) ? "tok-kw" : "";
      if (m[4]) return "tok-num";
      return "";
    });
  }

  function highlightJSON(src) {
    return tokenize(src, JSON_RE, (m, after) => {
      if (m[1]) {
        // A string followed by a colon is an object key.
        let k = after;
        while (k < src.length && /\s/.test(src[k])) k++;
        return src[k] === ":" ? "tok-typ" : "tok-str";
      }
      if (m[2]) return "tok-num";
      if (m[3]) return "tok-kw";
      return "";
    });
  }

  function highlight(code, lang) {
    switch (String(lang || "").toLowerCase()) {
      case "go": return highlightGo(code);
      case "bash": case "sh": case "shell": return highlightBash(code);
      case "sql": return highlightSQL(code);
      case "json": return highlightJSON(code);
      default: return escapeHtml(code);
    }
  }

  // Inline formatting. Code spans are extracted first and put back at the end,
  // so emphasis markers inside `code` cannot leak out, and emphasis that spans
  // a code span (e.g. "**every `.go` file**") still works.
  function inline(s) {
    const codes = [];
    let t = String(s).replace(/`([^`]+)`/g, (m, c) => {
      codes.push(c);
      return "\u0000" + (codes.length - 1) + "\u0000";
    });

    t = escapeHtml(t);
    t = t.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
    t = t.replace(/(^|[^*])\*([^*\n]+)\*/g, "$1<em>$2</em>");
    t = t.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, '<a href="$2" target="_blank" rel="noreferrer">$1</a>');

    return t.replace(/\u0000(\d+)\u0000/g, (m, n) => "<code>" + escapeHtml(codes[+n]) + "</code>");
  }

  // Split a markdown table row on pipes, ignoring pipes that are inside a code
  // span (|`a || b`|) or escaped (\|).
  function splitRow(line) {
    let t = line.trim();
    if (t.startsWith("|")) t = t.slice(1);
    if (t.endsWith("|")) t = t.slice(0, -1);

    const cells = [];
    let cur = "";
    let inCode = false;

    for (let k = 0; k < t.length; k++) {
      const ch = t[k];
      if (ch === "\\" && t[k + 1] === "|") { cur += "|"; k++; continue; }
      if (ch === "`") { inCode = !inCode; cur += ch; continue; }
      if (ch === "|" && !inCode) { cells.push(cur.trim()); cur = ""; continue; }
      cur += ch;
    }
    cells.push(cur.trim());
    return cells;
  }

  function isTableSep(line) {
    return !!line && line.indexOf("|") >= 0 && /^[\s|:\-]+$/.test(line) && line.indexOf("-") >= 0;
  }

  // A list item, with its indentation, so we can build nested lists.
  const LIST_RE = /^(\s*)([-*]|\d+[.)])\s+(.*)$/;
  const BLOCK_START_RE = /^\s*(#{1,6}\s|>|```|---\s*$|\*\*\*\s*$|___\s*$)/;

  // Collect one whole list block. A line that is indented and is not itself a
  // new item or the start of another block is a *continuation* of the previous
  // item — that is what lets a list item wrap over several source lines.
  function collectListItems(lines, start) {
    const items = [];
    let i = start;

    while (i < lines.length) {
      const m = lines[i].match(LIST_RE);
      if (m) {
        items.push({
          indent: m[1].replace(/\t/g, "    ").length,
          ordered: /\d/.test(m[2]),
          text: m[3].trim(),
        });
        i++;
        continue;
      }

      const line = lines[i];
      const blank = /^\s*$/.test(line);
      const indented = /^\s/.test(line);

      if (!blank && indented && !BLOCK_START_RE.test(line) && items.length) {
        items[items.length - 1].text += " " + line.trim();
        i++;
        continue;
      }
      break;
    }
    return { items, next: i };
  }

  // Turn the flat item list into a tree using indent levels.
  function buildList(items, start, indent) {
    const ordered = items[start].ordered;
    const nodes = [];
    let i = start;

    while (i < items.length && items[i].indent >= indent) {
      if (items[i].indent > indent) break;

      const node = { text: items[i].text, children: [] };
      i++;

      if (i < items.length && items[i].indent > indent) {
        const sub = buildList(items, i, items[i].indent);
        node.children.push(sub.node);
        i = sub.next;
      }
      nodes.push(node);
    }
    return { node: { ordered, items: nodes }, next: i };
  }

  function renderList(node) {
    const tag = node.ordered ? "ol" : "ul";
    return "<" + tag + ">" + node.items.map((it) =>
      "<li>" + inline(it.text) + it.children.map(renderList).join("") + "</li>"
    ).join("") + "</" + tag + ">";
  }

  function md(src) {
    const lines = String(src == null ? "" : src).replace(/\r\n/g, "\n").split("\n");
    const out = [];
    let i = 0;
    let para = [];

    const flushPara = () => {
      if (para.length) { out.push("<p>" + inline(para.join(" ")) + "</p>"); para = []; }
    };

    while (i < lines.length) {
      const line = lines[i];

      const fence = line.match(/^\s*```([A-Za-z0-9_+-]*)\s*$/);
      if (fence) {
        flushPara();
        const body = [];
        i++;
        while (i < lines.length && !/^\s*```\s*$/.test(lines[i])) { body.push(lines[i]); i++; }
        i++;
        out.push('<pre><code class="lang-' + fence[1] + '">' + highlight(body.join("\n"), fence[1]) + "</code></pre>");
        continue;
      }

      if (/^\s*$/.test(line)) { flushPara(); i++; continue; }

      const heading = line.match(/^(#{1,6})\s+(.*)$/);
      if (heading) {
        flushPara();
        const lvl = heading[1].length;
        out.push("<h" + lvl + ">" + inline(heading[2]) + "</h" + lvl + ">");
        i++; continue;
      }

      if (/^\s*(---|\*\*\*|___)\s*$/.test(line)) { flushPara(); out.push("<hr>"); i++; continue; }

      if (/^>\s?/.test(line)) {
        flushPara();
        const body = [];
        while (i < lines.length && /^>\s?/.test(lines[i])) { body.push(lines[i].replace(/^>\s?/, "")); i++; }
        out.push("<blockquote>" + md(body.join("\n")) + "</blockquote>");
        continue;
      }

      if (/\|/.test(line) && isTableSep(lines[i + 1])) {
        flushPara();
        const head = splitRow(line);
        i += 2;
        const rows = [];
        while (i < lines.length && /\|/.test(lines[i]) && !/^\s*$/.test(lines[i])) { rows.push(splitRow(lines[i])); i++; }
        out.push(
          '<div class="table-wrap"><table><thead><tr>' + head.map((c) => "<th>" + inline(c) + "</th>").join("") + "</tr></thead><tbody>" +
          rows.map((r) => "<tr>" + r.map((c) => "<td>" + inline(c) + "</td>").join("") + "</tr>").join("") +
          "</tbody></table></div>"
        );
        continue;
      }

      const match = line.match(LIST_RE);
      if (match) {
        flushPara();
        const collected = collectListItems(lines, i);
        const tree = buildList(collected.items, 0, collected.items[0].indent);
        out.push(renderList(tree.node));
        i = collected.next;
        continue;
      }

      para.push(line.trim());
      i++;
    }

    flushPara();
    return out.join("\n");
  }

  // -------------------------------------------------------------- helpers

  function stepsFor(topic) {
    const steps = [{ kind: "theory", key: topic.id + ":theory", title: "Theory" }];
    (topic.quiz || []).forEach((q, idx) => {
      steps.push({ kind: "quiz", key: topic.id + ":q:" + q.id, q: q, title: "Question " + (idx + 1) });
    });
    (topic.exercises || []).forEach((ex) => {
      steps.push({ kind: "exercise", key: topic.id + ":e:" + ex.id, ex: ex, title: ex.title });
    });
    return steps;
  }

  function buildFlat() {
    state.flat = [];
    state.curriculum.forEach((mod) => {
      (mod.topics || []).forEach((topic) => {
        stepsFor(topic).forEach((step) => {
          state.flat.push({ module: mod, topic: topic, step: step });
        });
      });
    });
  }

  function isDone(key) { return !!state.progress[key]; }

  function markDone(key, value) {
    if (value === false) delete state.progress[key];
    else state.progress[key] = true;
    write(LS_PROGRESS, state.progress);
    renderSidebar();
    renderProgress();
  }

  function topicDone(topic) {
    return stepsFor(topic).every((s) => isDone(s.key));
  }

  function topicCompletion(topic) {
    const steps = stepsFor(topic);
    return steps.filter((s) => isDone(s.key)).length + "/" + steps.length;
  }

  function currentCode(ex) {
    return state.codes[ex.id] != null ? state.codes[ex.id] : ex.starter;
  }

  function setCode(id, code) {
    state.codes[id] = code;
    write(LS_CODES, state.codes);
  }

  // ------------------------------------------------------------- sidebar

  function renderSidebar() {
    const host = $("#curriculum");
    host.innerHTML = "";

    if (!state.curriculum.length) {
      host.innerHTML = '<div class="loading">Curriculum is empty.</div>';
      return;
    }

    const activeTopic = state.flat[state.index] ? state.flat[state.index].topic.id : null;

    state.curriculum.forEach((mod) => {
      const wrap = document.createElement("div");
      wrap.className = "module";

      const doneTopics = mod.topics.filter(topicDone).length;
      const head = document.createElement("div");
      head.className = "module-head";
      head.innerHTML =
        '<span class="chev">▾</span><span>' + escapeHtml(mod.title) + "</span>" +
        '<span class="module-count">' + doneTopics + "/" + mod.topics.length + "</span>";
      head.addEventListener("click", () => wrap.classList.toggle("collapsed"));
      wrap.appendChild(head);

      const list = document.createElement("ul");
      list.className = "topic-list";
      mod.topics.forEach((topic) => {
        const li = document.createElement("li");
        li.className = "topic-item" + (topic.id === activeTopic ? " active" : "") + (topicDone(topic) ? " done" : "");
        li.innerHTML = '<span class="dot">✓</span><span>' + escapeHtml(topic.title) + "</span>";
        li.addEventListener("click", () => {
          const idx = state.flat.findIndex((f) => f.topic.id === topic.id);
          if (idx >= 0) goto(firstIncompleteIn(topic, idx));
          document.body.classList.remove("nav-open");
        });
        list.appendChild(li);
      });
      wrap.appendChild(list);
      host.appendChild(wrap);
    });
  }

  function firstIncompleteIn(topic, baseIndex) {
    const steps = stepsFor(topic);
    for (let i = 0; i < steps.length; i++) {
      if (!isDone(steps[i].key)) return baseIndex + i;
    }
    return baseIndex;
  }

  function renderProgress() {
    const total = state.flat.length || 1;
    const done = state.flat.filter((f) => isDone(f.step.key)).length;
    const pct = Math.round((done / total) * 100);
    $("#progress-fill").style.width = pct + "%";
    $("#progress-label").textContent = pct + "%";
  }

  // ------------------------------------------------------------ navigation

  function goto(index) {
    state.index = Math.max(0, Math.min(state.flat.length - 1, index));
    render();
  }

  let lastRenderedIndex = -1;

  function render() {
    renderSidebar();
    renderProgress();

    const entry = state.flat[state.index];
    const content = $("#content");
    if (!entry) {
      content.innerHTML = '<div class="empty-state"><h1>Nothing to show</h1></div>';
      return;
    }

    const keepScroll = state.index === lastRenderedIndex ? content.scrollTop : 0;
    lastRenderedIndex = state.index;

    const host = document.createElement("div");
    host.className = "wrap";
    content.innerHTML = "";
    content.appendChild(host);

    // Page header: the title appears once, with the step navigation beside it.
    const header = document.createElement("header");
    header.className = "page-header";

    const menuBtn = document.createElement("button");
    menuBtn.className = "icon-btn only-mobile";
    menuBtn.setAttribute("aria-label", "Toggle navigation");
    menuBtn.textContent = "☰";
    menuBtn.addEventListener("click", () => document.body.classList.toggle("nav-open"));

    // The heading group keeps the title and its completion tick together, and
    // the fixed min-height stops the row resizing as steps change.
    const heading = document.createElement("div");
    heading.className = "page-heading";

    const title = document.createElement("h1");
    title.className = "page-title";
    title.textContent = entry.topic.title;
    heading.appendChild(title);

    if (topicDone(entry.topic)) {
      const done = document.createElement("span");
      done.className = "title-check";
      done.textContent = "✓";
      done.title = "Topic complete";
      done.setAttribute("aria-label", "Topic complete");
      heading.appendChild(done);
    }

    const nav = document.createElement("div");
    nav.className = "page-nav";

    const prev = document.createElement("button");
    prev.className = "ghost-btn";
    prev.textContent = "← Previous";
    prev.disabled = state.index === 0;
    prev.addEventListener("click", () => goto(state.index - 1));

    const next = document.createElement("button");
    next.className = "ghost-btn";
    next.textContent = "Next →";
    next.disabled = state.index === state.flat.length - 1;
    next.addEventListener("click", () => {
      if (entry.step.kind === "theory") markDone(entry.step.key);
      goto(state.index + 1);
    });

    nav.appendChild(prev);
    nav.appendChild(next);

    header.appendChild(menuBtn);
    header.appendChild(heading);
    header.appendChild(nav);
    host.appendChild(header);

    host.appendChild(renderStepTabs(entry));

    if (entry.step.kind === "theory") host.appendChild(renderTheory(entry));
    if (entry.step.kind === "quiz") host.appendChild(renderQuiz(entry));
    if (entry.step.kind === "exercise") host.appendChild(renderExercise(entry));

    content.scrollTop = keepScroll;
  }

  function renderStepTabs(entry) {
    const steps = stepsFor(entry.topic);
    const base = state.flat.findIndex((f) => f.topic.id === entry.topic.id);
    const host = document.createElement("div");
    host.className = "step-tabs";

    steps.forEach((step, i) => {
      const btn = document.createElement("button");
      btn.className = "step-tab" + (step.key === entry.step.key ? " active" : "");
      btn.innerHTML = (isDone(step.key) ? '<span class="tick">✓</span>' : "") + escapeHtml(step.title);
      btn.addEventListener("click", () => goto(base + i));
      host.appendChild(btn);
    });
    return host;
  }

  function theoryBody(text) {
    const card = document.createElement("div");
    card.className = "card";
    card.innerHTML = '<div class="theory">' + md(text) + "</div>";
    return card;
  }

  // ---------------------------------------------------------------- theory

  function renderTheory(entry) {
    const frag = document.createDocumentFragment();

    frag.appendChild(theoryBody(entry.topic.theory));

    const row = document.createElement("div");
    row.className = "row";
    const btn = document.createElement("button");
    btn.className = "primary-btn next-btn";
    btn.textContent = "Next →";
    btn.addEventListener("click", () => {
      markDone(entry.step.key);
      goto(state.index + 1);
    });
    row.appendChild(btn);
    frag.appendChild(row);

    return frag;
  }

  // ------------------------------------------------------------------ quiz

  function renderQuiz(entry) {
    const q = entry.step.q;
    const key = entry.step.key;
    const qstate = state.quiz[key] || (state.quiz[key] = { selected: new Set(), checked: false });

    const frag = document.createDocumentFragment();
    const card = document.createElement("div");
    card.className = "card";

    const multi = !!q.multi;
    const meta = document.createElement("div");
    meta.className = "q-meta";
    meta.textContent = multi ? "Select all that apply" : "Select one answer";
    card.appendChild(meta);

    const prompt = document.createElement("div");
    prompt.className = "q-prompt";
    prompt.innerHTML = inline(q.prompt || "");
    card.appendChild(prompt);

    if (q.code) {
      // The wrapper carries .theory so the standard .theory pre/code rules apply;
      // putting .theory on the <pre> itself would leave it unstyled and let the
      // inline-code rules (white-space: nowrap) flatten the whole block.
      const holder = document.createElement("div");
      holder.className = "theory q-code";
      const pre = document.createElement("pre");
      const code = document.createElement("code");
      code.innerHTML = highlight(q.code, "go");
      pre.appendChild(code);
      holder.appendChild(pre);
      card.appendChild(holder);
    }

    const opts = document.createElement("div");
    opts.className = "options";

    let submitBtn = null;

    (q.options || []).forEach((opt, i) => {
      const label = document.createElement("label");
      label.className = "option";
      if (qstate.checked) {
        label.classList.add("locked");
        if (opt.correct) label.classList.add("correct");
        else if (qstate.selected.has(i)) label.classList.add("wrong");
      }

      const input = document.createElement("input");
      input.type = multi ? "checkbox" : "radio";
      input.name = "q-" + q.id;
      input.checked = qstate.selected.has(i);
      input.disabled = qstate.checked;
      input.addEventListener("change", () => {
        if (multi) {
          if (input.checked) qstate.selected.add(i); else qstate.selected.delete(i);
        } else {
          qstate.selected = new Set([i]);
        }
        if (submitBtn) submitBtn.disabled = qstate.selected.size === 0;
      });

      const body = document.createElement("span");
      body.innerHTML = inline(opt.text);
      if (qstate.checked && opt.rationale) {
        body.innerHTML += '<span class="rationale">' + inline(opt.rationale) + "</span>";
      }

      label.appendChild(input);
      label.appendChild(body);
      opts.appendChild(label);
    });

    card.appendChild(opts);

    if (qstate.checked) {
      const correct = isAnswerCorrect(q, qstate.selected);
      const verdict = document.createElement("div");
      verdict.className = "verdict " + (correct ? "ok" : "no");
      verdict.innerHTML = correct
        ? '<span class="mark">✓</span> Correct'
        : '<span class="mark">✕</span> Not quite — see the highlighted answers below';
      card.appendChild(verdict);

      if (q.explanation) {
        const ex = document.createElement("div");
        ex.className = "explain";
        ex.innerHTML = "<strong>Why:</strong> " + inline(q.explanation);
        card.appendChild(ex);
      }
    }

    const row = document.createElement("div");
    row.className = "row";

    if (!qstate.checked) {
      const check = document.createElement("button");
      check.className = "primary-btn";
      check.textContent = "Submit answer";
      check.disabled = qstate.selected.size === 0;
      submitBtn = check;
      check.addEventListener("click", () => {
        qstate.checked = true;
        if (isAnswerCorrect(q, qstate.selected)) markDone(key);
        render();
      });
      row.appendChild(check);
    } else {
      const retry = document.createElement("button");
      retry.className = "ghost-btn";
      retry.textContent = "Try again";
      retry.addEventListener("click", () => {
        state.quiz[key] = { selected: new Set(), checked: false };
        render();
      });
      row.appendChild(retry);

      const next = document.createElement("button");
      next.className = "primary-btn next-btn";
      next.textContent = "Next →";
      // Moving on does NOT complete the question — only a correct answer does.
      next.addEventListener("click", () => goto(state.index + 1));
      row.appendChild(next);
    }

    card.appendChild(row);
    frag.appendChild(card);
    return frag;
  }

  function isAnswerCorrect(q, selected) {
    const correct = new Set();
    (q.options || []).forEach((o, i) => { if (o.correct) correct.add(i); });
    if (correct.size !== selected.size) return false;
    for (const i of correct) if (!selected.has(i)) return false;
    return true;
  }

  // -------------------------------------------------------------- exercise

  function renderExercise(entry) {
    const ex = entry.step.ex;
    const key = entry.step.key;

    const frag = document.createDocumentFragment();

    // ---- question card: the prompt, plus the hint / solution controls. Keeping
    // those controls with the question is what makes them findable.
    const hints = ex.hints || [];
    const revealed = state.hints[ex.id] || 0;
    const solutionOpen = !!state.solutions[ex.id];

    const brief = theoryBody(ex.prompt);

    const controls = document.createElement("div");
    controls.className = "row question-actions";

    if (hints.length) {
      // Always render the button, even once every hint is out: removing it made
      // the "Show solution" button jump left, and a disabled control also tells
      // the learner there is nothing left to reveal.
      const exhausted = revealed >= hints.length;
      const hb = document.createElement("button");
      hb.className = "ghost-btn";
      hb.disabled = exhausted;
      if (exhausted) {
        hb.textContent = "All hints shown";
        hb.title = "Every hint for this exercise is already on screen";
      } else {
        hb.textContent = revealed === 0
          ? "Reveal a hint"
          : "Another hint (" + revealed + "/" + hints.length + ")";
        hb.addEventListener("click", () => {
          state.hints[ex.id] = (state.hints[ex.id] || 0) + 1;
          render();
        });
      }
      controls.appendChild(hb);
    }

    if (ex.solution) {
      const sb = document.createElement("button");
      sb.className = "ghost-btn";
      sb.textContent = solutionOpen ? "Hide solution" : "Show solution";
      sb.addEventListener("click", () => {
        if (solutionOpen) {
          state.solutions[ex.id] = false;
          render();
          return;
        }
        if (confirm("Reveal the reference solution? Try Submit first if you can.")) {
          state.solutions[ex.id] = true;
          render();
        }
      });
      controls.appendChild(sb);
    }

    if (controls.children.length) brief.appendChild(controls);

    if (revealed > 0) {
      const hintWrap = document.createElement("div");
      hintWrap.className = "hints";
      for (let i = 0; i < Math.min(revealed, hints.length); i++) {
        const h = document.createElement("div");
        h.className = "hint";
        h.innerHTML = "<strong>Hint " + (i + 1) + ":</strong> " + inline(hints[i]);
        hintWrap.appendChild(h);
      }
      brief.appendChild(hintWrap);
    }

    if (ex.solution && solutionOpen) {
      const sol = document.createElement("div");
      sol.className = "solution";
      sol.innerHTML = '<pre><code class="lang-go">' + highlight(ex.solution, "go") + "</code></pre>";
      brief.appendChild(sol);
    }

    frag.appendChild(brief);

    // ---- editor card
    const card = document.createElement("div");
    card.className = "card";

    const shell = document.createElement("div");
    shell.className = "editor-shell";

    // A chrome bar with the file name and reset, so the editor reads as an
    // editor rather than as another code block.
    const bar = document.createElement("div");
    bar.className = "editor-bar";

    const fileName = document.createElement("span");
    fileName.className = "file-name";
    fileName.textContent = "main.go";

    const barSpacer = document.createElement("span");
    barSpacer.className = "spacer";

    let runProgram = null;
    const runProgramBtn = document.createElement("button");
    runProgramBtn.className = "editor-action-btn run-program-btn";
    runProgramBtn.dataset.action = "run";
    runProgramBtn.textContent = "▶ Run program";
    runProgramBtn.title = "go run . — compiles and executes main()";
    runProgramBtn.addEventListener("click", () => {
      if (runProgram) runProgram();
    });

    const submitBtn = document.createElement("button");
    submitBtn.className = "editor-action-btn submit-btn";
    submitBtn.dataset.action = "submit";
    submitBtn.textContent = "Submit";
    submitBtn.title = "Run the graded tests; passing marks this exercise complete";

    const resetBtn = document.createElement("button");
    resetBtn.className = "editor-action-btn reset-btn";
    resetBtn.textContent = "Reset code";
    resetBtn.addEventListener("click", () => {
      if (confirm("Discard your changes and restore the starter code?")) {
        editor.value = ex.starter;
        setCode(ex.id, ex.starter);
        refreshHighlight();
        syncScroll();
      }
    });

    bar.appendChild(fileName);
    bar.appendChild(barSpacer);
    bar.appendChild(runProgramBtn);
    bar.appendChild(submitBtn);
    bar.appendChild(resetBtn);
    shell.appendChild(bar);

    const body = document.createElement("div");
    body.className = "editor-body";

    // A highlighted copy of the code sits *behind* the textarea. The textarea's
    // own text is transparent; only its caret and selection show through.
    // The two must share identical font metrics, padding and white-space.
    const highlightPre = document.createElement("pre");
    highlightPre.className = "editor-highlight";
    highlightPre.setAttribute("aria-hidden", "true");
    const highlightCode = document.createElement("code");
    highlightPre.appendChild(highlightCode);

    const editor = document.createElement("textarea");
    editor.className = "editor";
    editor.spellcheck = false;
    editor.setAttribute("wrap", "off"); // no soft wrapping, or the overlay drifts
    editor.value = currentCode(ex);

    const refreshHighlight = () => {
      highlightCode.innerHTML = highlight(editor.value, "go") + "\n";
      fitOverlay();
    };

    const syncScroll = () => {
      highlightPre.scrollTop = editor.scrollTop;
      highlightPre.scrollLeft = editor.scrollLeft;
      fitOverlay();
    };

    // Match the textarea's *content* box exactly. On platforms with classic
    // (non-overlay) scrollbars the textarea's content area is smaller than the
    // overlay's, which would otherwise leave a strip of phantom text showing.
    const fitOverlay = () => {
      const h = editor.clientHeight;
      highlightPre.style.height = h ? h + "px" : "100%";
    };

    refreshHighlight();

    if (typeof ResizeObserver !== "undefined") {
      new ResizeObserver(fitOverlay).observe(editor);
    }

    editor.addEventListener("input", () => {
      setCode(ex.id, editor.value);
      refreshHighlight();
    });
    editor.addEventListener("scroll", syncScroll);
    editor.addEventListener("keydown", (e) => onEditorKey(e, editor, ex, refreshHighlight));

    body.appendChild(highlightPre);
    body.appendChild(editor);
    shell.appendChild(body);
    card.appendChild(shell);

    const stdin = document.createElement("textarea");
    stdin.className = "stdin-box";
    stdin.placeholder = "stdin (optional) — used when you run your program";
    stdin.value = ex.sampleIn || "";
    card.appendChild(stdin);

    const status = document.createElement("div");
    status.className = "run-status";
    status.setAttribute("role", "status");
    status.setAttribute("aria-live", "polite");
    card.appendChild(status);

    const buttons = { runBtn: runProgramBtn, submitBtn: submitBtn };
    runProgram = () => doExecute(entry, "run", editor.value, stdin.value, buttons, status);
    submitBtn.addEventListener("click", () => doExecute(entry, "test", editor.value, stdin.value, buttons, status));

    // ---- output: part of the editor card, so the results appear right under the
    // code that produced them instead of on a separate island.
    const out = state.output[ex.id];
    if (out) {
      const outputEl = renderOutput(entry, out);
      card.appendChild(outputEl);
      if (out.mode === "test" && typeof requestAnimationFrame !== "undefined") {
        requestAnimationFrame(() => {
          const banner = outputEl.querySelector(".result-banner");
          if (banner) banner.scrollIntoView({ block: "nearest" });
        });
      }
    }

    frag.appendChild(card);

    const nextRow = document.createElement("div");
    nextRow.className = "exercise-next-row";
    const nextBtn = document.createElement("button");
    nextBtn.className = "primary-btn next-btn";
    nextBtn.textContent = "Next →";
    nextBtn.disabled = state.index === state.flat.length - 1;
    nextBtn.addEventListener("click", () => goto(state.index + 1));
    nextRow.appendChild(nextBtn);
    frag.appendChild(nextRow);

    // The editor is usually built while still detached from the document, so
    // clientHeight is 0 at that point. Re-fit once the browser has laid it out.
    if (typeof requestAnimationFrame !== "undefined") requestAnimationFrame(fitOverlay);

    return frag;
  }

  function onEditorKey(e, editor, ex, refresh) {
    if (e.key === "Tab") {
      e.preventDefault();
      const start = editor.selectionStart;
      const end = editor.selectionEnd;
      const value = editor.value;

      if (start !== end && value.slice(start, end).includes("\n")) {
        const from = value.lastIndexOf("\n", start - 1) + 1;
        const block = value.slice(from, end);
        const shifted = block.split("\n").map((l) => (e.shiftKey ? l.replace(/^\t/, "") : "\t" + l)).join("\n");
        editor.value = value.slice(0, from) + shifted + value.slice(end);
        editor.selectionStart = from;
        editor.selectionEnd = from + shifted.length;
      } else {
        editor.value = value.slice(0, start) + "\t" + value.slice(end);
        editor.selectionStart = editor.selectionEnd = start + 1;
      }
      setCode(ex.id, editor.value);
      if (refresh) refresh();
      return;
    }

    if (e.key === "Enter" && !e.shiftKey && !e.ctrlKey && !e.metaKey) {
      const start = editor.selectionStart;
      const value = editor.value;
      const lineStart = value.lastIndexOf("\n", start - 1) + 1;
      const line = value.slice(lineStart, start);
      const indent = (line.match(/^\s*/) || [""])[0];
      const extra = /[{[(]\s*$/.test(line) ? "\t" : "";
      if (indent || extra) {
        e.preventDefault();
        const insert = "\n" + indent + extra;
        editor.value = value.slice(0, start) + insert + value.slice(start);
        editor.selectionStart = editor.selectionEnd = start + insert.length;
        setCode(ex.id, editor.value);
        if (refresh) refresh();
      }
      return;
    }

    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
      e.preventDefault();
      const card = editor.closest(".card");
      const sel = e.shiftKey ? '[data-action="run"]' : '[data-action="submit"]';
      const btn = card && card.querySelector(sel);
      if (btn && !btn.disabled) btn.click();
    }
  }

  async function doExecute(entry, mode, code, stdin, buttons, status) {
    const ex = entry.step.ex;
    const active = [buttons.submitBtn, buttons.runBtn].filter(Boolean);
    active.forEach((btn) => { btn.disabled = true; });

    const busy = mode === "run" ? "Compiling and running…" : "Running the graded tests…";
    status.innerHTML = '<span class="spinner"></span> ' + busy;

    let data;
    try {
      const res = await fetch("/api/execute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          exerciseId: ex.id,
          files: [{ name: "main.go", content: code }],
          mode: mode,
          stdin: stdin || "",
        }),
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || ("HTTP " + res.status));
      }
      data = await res.json();
    } catch (err) {
      data = {
        ok: false,
        stdout: "",
        stderr: String(err && err.message ? err.message : err),
        exitCode: -1,
        tests: [],
        networkError: true,
      };
    }

    let tab = "stdout";
    if (mode === "test") tab = "tests";
    if (data.stderr && !data.stdout && (!data.tests || !data.tests.length)) tab = "stderr";

    state.output[ex.id] = { tab: tab, data: data, mode: mode };

    active.forEach((btn) => { btn.disabled = false; });

    if (mode === "test" && data.ok) {
      markDone(entry.step.key, true);
    }
    status.textContent = data.timedOut ? "Timed out" : "";
    render();
  }

  function renderOutput(entry, out) {
    const data = out.data || {};
    const tests = data.tests || [];
    const passed = tests.filter((t) => t.pass).length;
    const failed = tests.length - passed;
    const firstFailure = tests.find((t) => !t.pass);

    const wrap = document.createElement("div");
    wrap.className = "output";

    // The verdict is the first thing under the buttons. Previously the only
    // signal was a small badge in the output header, which was easy to miss.
    if (out.mode === "test") {
      const banner = document.createElement("div");
      banner.className = "result-banner " + (data.ok ? "ok" : "no");
      banner.setAttribute("role", "status");
      banner.setAttribute("aria-live", "polite");

      const icon = document.createElement("span");
      icon.className = "result-icon";
      icon.setAttribute("aria-hidden", "true");
      icon.textContent = data.ok ? "✓" : "✕";

      const text = document.createElement("div");
      text.className = "result-text";

      const title = document.createElement("div");
      title.className = "result-title";

      const sub = document.createElement("div");
      sub.className = "result-sub";

      if (data.networkError) {
        title.textContent = "Could not reach the runner";
        sub.textContent = data.stderr || "Check that the GoTut server is still running, then try again.";
      } else if (data.timedOut) {
        title.textContent = "Timed out";
        sub.textContent = "Your program ran for too long. Check for an infinite loop, then try again.";
      } else if (data.ok) {
        title.textContent = "All tests passed";
        if (tests.length) {
          sub.textContent = passed + "/" + tests.length + " tests passed. This exercise is complete.";
        } else {
          sub.textContent = "The package built successfully. This exercise is complete.";
        }
      } else if (tests.length) {
        title.textContent = failed + " of " + tests.length + " tests failed";
        sub.textContent = firstFailure
          ? "First failure: " + firstFailure.name
          : "Open the Tests tab to see the details.";
      } else {
        title.textContent = "Build failed";
        sub.textContent = "Fix the compiler error, then press Submit again. See the Errors tab.";
      }

      text.appendChild(title);
      text.appendChild(sub);
      banner.appendChild(icon);
      banner.appendChild(text);
      wrap.appendChild(banner);
    }

    const head = document.createElement("div");
    head.className = "output-head";

    const tabs = [
      { id: "tests", label: "Tests" },
      { id: "stdout", label: "Output" },
      { id: "stderr", label: "Errors" },
    ];

    const body = document.createElement("div");

    const draw = (tab) => {
      out.tab = tab;
      Array.from(head.querySelectorAll(".output-tab")).forEach((t) => {
        t.classList.toggle("active", t.dataset.tab === tab);
      });
      body.innerHTML = "";
      body.className = "";

      if (tab === "tests") {
        if (!tests.length) {
          const p = document.createElement("pre");
          p.className = "output-body empty";
          p.textContent = data.stderr
            ? "No tests ran — the package probably failed to build. See the Errors tab."
            : "No tests were run. Press Submit to run the graded tests.";
          body.appendChild(p);
        } else {
          tests.forEach((t) => {
            const line = document.createElement("div");
            line.className = "test-line " + (t.pass ? "pass" : "fail");
            line.innerHTML = '<span class="mark">' + (t.pass ? "PASS" : "FAIL") + '</span><span class="test-name">' + escapeHtml(t.name) + "</span>";
            if (!t.pass && t.output) {
              const detail = document.createElement("pre");
              detail.className = "test-output";
              detail.textContent = t.output;
              line.appendChild(detail);
            }
            body.appendChild(line);
          });
        }
        return;
      }

      const text = tab === "stdout" ? data.stdout : data.stderr;
      const p = document.createElement("pre");
      p.className = "output-body" + (text ? "" : " empty");
      p.textContent = text || (tab === "stdout" ? "No output on stdout." : "No errors.");
      body.appendChild(p);
    };

    tabs.forEach((t) => {
      const btn = document.createElement("button");
      btn.className = "output-tab";
      btn.dataset.tab = t.id;
      btn.textContent = t.label;
      const n = (data.tests || []).length;
      if (t.id === "tests" && n) btn.textContent += " (" + n + ")";
      btn.addEventListener("click", () => draw(t.id));
      head.appendChild(btn);
    });

    const spacer = document.createElement("span");
    spacer.className = "spacer";
    head.appendChild(spacer);

    if (out.mode === "test") {
      const badge = document.createElement("span");
      const n = (data.tests || []).length;
      const passed = (data.tests || []).filter((t) => t.pass).length;
      if (data.ok) { badge.className = "badge ok"; badge.textContent = n ? passed + "/" + n + " passed" : "ok"; }
      else if (n) { badge.className = "badge no"; badge.textContent = passed + "/" + n + " passed"; }
      else { badge.className = "badge no"; badge.textContent = "build failed"; }
      head.appendChild(badge);
    } else {
      const badge = document.createElement("span");
      if (data.ok) { badge.className = "badge ok"; badge.textContent = "exit 0"; }
      else if (data.timedOut) { badge.className = "badge warn"; badge.textContent = "timeout"; }
      else { badge.className = "badge no"; badge.textContent = "exit " + data.exitCode; }
      head.appendChild(badge);
    }

    const time = document.createElement("span");
    time.className = "muted";
    time.style.fontSize = "11.5px";
    time.style.marginLeft = "8px";
    time.textContent = (data.durationMs || 0) + " ms";
    head.appendChild(time);

    wrap.appendChild(head);
    wrap.appendChild(body);
    draw(out.tab || "stdout");
    return wrap;
  }

  // ----------------------------------------------------------------- boot

  // Two exercises moved topics when types and zero values were split apart.
  // Their ids (and therefore saved code) are unchanged, but progress keys
  // include the topic id, so carry completion across to the new location.
  const PROGRESS_MOVES = {
    "m1t2:e:m1t2e1": "m1t3:e:m1t2e1",
    "m1t3:e:m1t3e3": "m1t5:e:m1t3e3",
  };

  function migrateProgress() {
    let changed = false;
    Object.keys(PROGRESS_MOVES).forEach((oldKey) => {
      if (!state.progress[oldKey]) return;
      const newKey = PROGRESS_MOVES[oldKey];
      if (!state.progress[newKey]) state.progress[newKey] = true;
      delete state.progress[oldKey];
      changed = true;
    });
    if (changed) write(LS_PROGRESS, state.progress);
  }

  function applyTheme(theme) {
    document.documentElement.setAttribute("data-theme", theme);
    write(LS_THEME, theme);
  }

  function init() {
    migrateProgress();

    const savedTheme = read(LS_THEME, "dark");
    applyTheme(typeof savedTheme === "string" ? savedTheme : "dark");

    $("#theme-toggle").addEventListener("click", () => {
      const now = document.documentElement.getAttribute("data-theme");
      applyTheme(now === "dark" ? "light" : "dark");
    });

    $("#scrim").addEventListener("click", () => document.body.classList.remove("nav-open"));

    $("#reset-progress").addEventListener("click", () => {
      if (!confirm("Clear all progress and saved code?")) return;
      state.progress = {};
      state.codes = {};
      write(LS_PROGRESS, state.progress);
      write(LS_CODES, state.codes);
      render();
    });

    document.addEventListener("keydown", (e) => {
      if (e.target && /INPUT|TEXTAREA/.test(e.target.tagName)) return;
      if (e.key === "ArrowRight") goto(state.index + 1);
      if (e.key === "ArrowLeft") goto(state.index - 1);
    });

    fetch("/api/curriculum")
      .then((r) => {
        if (!r.ok) throw new Error("HTTP " + r.status);
        return r.json();
      })
      .then((data) => {
        state.curriculum = data || [];
        buildFlat();
        render();
      })
      .catch((err) => {
        $("#curriculum").innerHTML = '<div class="loading">Could not load the curriculum: ' + escapeHtml(err.message) + "</div>";
      });
  }

  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", init);
  else init();
})();
