(() => {
  const tokenKey = "trama.pb.token";
  let token = sessionStorage.getItem(tokenKey) || "";
  let account = null;
  let goals = [];
  let authMode = "login";

  const byId = (id) => document.getElementById(id);
  const authScreen = byId("auth-screen");
  const workspace = byId("workspace");
  const authForm = byId("auth-form");
  const authError = byId("auth-error");
  const notice = byId("notice");
  const authSubmit = byId("auth-submit");
  const authSwitch = byId("auth-switch");

  class PocketBaseError extends Error {
    constructor(message, status) {
      super(message);
      this.status = status;
    }
  }

  async function api(path, { method = "GET", body, authenticated = true } = {}) {
    const headers = { Accept: "application/json" };
    if (body !== undefined) headers["Content-Type"] = "application/json";
    if (authenticated && token) headers.Authorization = `Bearer ${token}`;
    const response = await fetch(path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      cache: "no-store",
    });
    const text = await response.text();
    let data = null;
    if (text) {
      try { data = JSON.parse(text); } catch { data = { message: text }; }
    }
    if (!response.ok) {
      const fieldError = data?.data && Object.values(data.data).find((item) => item?.message)?.message;
      throw new PocketBaseError(fieldError || data?.message || `La solicitud falló (${response.status}).`, response.status);
    }
    return data;
  }

  function saveAuth(data) {
    token = data.token;
    account = data.record;
    sessionStorage.setItem(tokenKey, token);
  }

  function clearAuth() {
    token = "";
    account = null;
    sessionStorage.removeItem(tokenKey);
  }

  function escapeHTML(value) {
    return String(value ?? "").replace(/[&<>"']/g, (char) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    })[char]);
  }

  function showNotice(message, isError = false) {
    notice.textContent = message;
    notice.classList.toggle("notice-error", isError);
    notice.hidden = !message;
  }

  function showAuthError(message) {
    authError.textContent = message;
    authError.hidden = !message;
  }

  function renderAuthMode() {
    const registering = authMode === "register";
    byId("auth-title").textContent = registering ? "Crea tu cuenta" : "Inicia sesión";
    byId("auth-hint").textContent = registering
      ? "Tu cuenta abre un espacio independiente con una ruta inicial lista. Aún no hay recuperación por correo; guarda tu contraseña."
      : "Entra a tu cuenta para continuar tu ruta de aprendizaje.";
    byId("password").autocomplete = registering ? "new-password" : "current-password";
    authSubmit.innerHTML = registering ? "Crear mi espacio <span aria-hidden=\"true\">→</span>" : "Iniciar sesión <span aria-hidden=\"true\">→</span>";
    byId("auth-switch-line").textContent = registering ? "¿Ya tienes una cuenta?" : "¿Primera vez en Trama?";
    authSwitch.textContent = registering ? "Volver a iniciar sesión" : "Crear una cuenta personal";
    showAuthError("");
  }

  function starterPath(domain, purpose, level) {
    return [
      {
        name: "Una meta que importa",
        summary: "Conecta el tema con algo que sí quieres lograr.",
        content: `Empieza por tu objetivo: ${purpose} Observa qué sabes ya y qué te gustaría comprender mejor.`,
        prompt: `¿Qué te gustaría comprender de ${domain} y por qué te importa?`,
        mastery: 0,
      },
      {
        name: "Ideas que se conectan",
        summary: "Busca relaciones entre conceptos, no solo definiciones.",
        content: `En el nivel ${level}, elige dos ideas de ${domain} y explica cómo una ayuda a entender la otra.`,
        prompt: `¿Qué dos ideas de ${domain} se relacionan y cómo?`,
        mastery: 0,
      },
      {
        name: "Una aplicación práctica",
        summary: "Lleva lo aprendido a un ejemplo cercano.",
        content: `Piensa en una situación cotidiana en la que podrías usar ${domain}. Describe un primer paso y qué aprenderías al intentarlo.`,
        prompt: `Describe una aplicación práctica de ${domain} vinculada con tu objetivo.`,
        mastery: 0,
      },
    ];
  }

  async function ensureInitialSession() {
    const response = await api("/api/collections/learning_goals/records?sort=-created&perPage=100");
    goals = Array.isArray(response.items) ? response.items : [];
    if (goals.length === 0) {
      const initial = await api("/api/collections/learning_goals/records", {
        method: "POST",
        body: {
          user: account.id,
          domain: "Aprender a aprender",
          level: "inicial",
          purpose: "Explorar Trama con una ruta privada de demostración.",
          minutes_per_week: 90,
          status: "active",
          progress: 0,
          path: starterPath("Aprender a aprender", "Explorar Trama con una ruta privada de demostración.", "inicial"),
        },
      });
      goals = [initial];
    }
    renderWorkspace();
  }

  function renderWorkspace() {
    authScreen.hidden = true;
    workspace.hidden = false;
    byId("account-controls").hidden = false;
    byId("account-email").textContent = account?.email || "Cuenta personal";
    byId("route-count").textContent = `${goals.length} ${goals.length === 1 ? "ruta" : "rutas"}`;
    byId("goal-list").innerHTML = goals.map((goal, index) => `
      <article class="route-card">
        <div class="route-card-top"><span class="card-kicker">${index === 0 ? "TU RUTA INICIAL" : "TU OBJETIVO"}</span><span class="route-private">Privada</span></div>
        <h3>${escapeHTML(goal.domain)}</h3>
        <p>${escapeHTML(goal.purpose)}</p>
        <div class="route-progress"><span style="width:${Math.max(0, Math.min(100, Number(goal.progress || 0) * 100))}%"></span></div>
        <button class="button button-secondary" type="button" data-open-goal="${escapeHTML(goal.id)}">Abrir mi ruta <span aria-hidden="true">→</span></button>
      </article>`).join("");
    if (!goals.some((goal) => goal.id === byId("goal-detail").dataset.goalId)) {
      byId("goal-detail").hidden = true;
      byId("goal-detail").replaceChildren();
    }
    showNotice("");
  }

  function renderGoal(goal) {
    const detail = byId("goal-detail");
    const path = Array.isArray(goal.path) ? goal.path : [];
    const progress = Math.max(0, Math.min(100, Number(goal.progress || 0) * 100));
    detail.dataset.goalId = goal.id;
    detail.innerHTML = `
      <div class="crumb"><span>RUTA DE APRENDIZAJE</span><span>${escapeHTML(goal.level)}</span></div>
      <section class="goal-heading"><div><p class="eyebrow"><span class="eyebrow-line"></span> TU OBJETIVO</p><h2>Aprender <em>${escapeHTML(goal.domain)}</em></h2><p class="intro compact">${escapeHTML(goal.purpose)}</p></div><div class="time-chip">◷ ${escapeHTML(goal.minutes_per_week)} min / semana</div></section>
      <section class="progress-panel" data-testid="progress"><div class="progress-top"><div><span class="card-kicker">TU AVANCE</span><strong>${progress.toFixed(0)}%</strong></div><span class="progress-copy">El dominio se construye paso a paso.</span></div><div class="progress-track" role="progressbar" aria-label="Progreso del objetivo" aria-valuemin="0" aria-valuemax="100" aria-valuenow="${progress.toFixed(0)}"><span style="width:${progress.toFixed(0)}%"></span></div></section>
      <section class="map-section"><div class="section-heading"><div><p class="eyebrow"><span class="eyebrow-line"></span> TU MAPA</p><h2>Una idea lleva a la siguiente.</h2></div><span class="map-count">${path.length} pasos</span></div>
        <ol class="concept-list">${path.map((concept, index) => `
          <li class="concept-card concept-card-expanded">
            <div class="concept-step"><span>${String(index + 1).padStart(2, "0")}</span></div>
            <div class="concept-copy"><h3>${escapeHTML(concept.name)}</h3><p>${escapeHTML(concept.summary)}</p><span class="concept-status">${Number(concept.mastery || 0) > 0 ? "PRÁCTICA COMPLETADA" : "POR EXPLORAR"}</span>
              <div class="lesson-body"><span class="lesson-label">PARA EMPEZAR</span><p>${escapeHTML(concept.content)}</p></div>
              ${concept.last_answer ? `<p class="saved-answer"><strong>Tu última respuesta:</strong> ${escapeHTML(concept.last_answer)}</p><p class="saved-feedback">${escapeHTML(concept.feedback || "")}</p>` : ""}
              <form class="goal-form attempt-form" data-goal-id="${escapeHTML(goal.id)}" data-index="${index}">
                <label for="answer-${index}">${escapeHTML(concept.prompt || "Tu explicación")}</label>
                <textarea id="answer-${index}" name="answer" rows="4" maxlength="4000" placeholder="Explica con tus propias palabras..." required></textarea>
                <button class="button button-primary" type="submit">Guardar respuesta y revisar <span aria-hidden="true">→</span></button>
              </form>
            </div>
          </li>`).join("")}</ol>
        <p class="prototype-footnote"><span aria-hidden="true">✳</span> Ruta inicial de demostración. La evaluación es heurística y no representa una revisión de IA.</p>
      </section>`;
    detail.hidden = false;
    detail.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  async function handleAuthSubmit(event) {
    event.preventDefault();
    showAuthError("");
    const email = byId("email").value.trim().toLowerCase();
    const password = byId("password").value;
    authSubmit.disabled = true;
    try {
      if (authMode === "register") {
        await api("/api/collections/users/records", {
          method: "POST",
          authenticated: false,
          body: { email, password, passwordConfirm: password, emailVisibility: false },
        });
      }
      const auth = await api("/api/collections/users/auth-with-password", {
        method: "POST",
        authenticated: false,
        body: { identity: email, password },
      });
      saveAuth(auth);
      await ensureInitialSession();
      authForm.reset();
    } catch (error) {
      showAuthError(authMode === "register" && error.status === 400
        ? "No se pudo crear la cuenta. Revisa el correo y usa una contraseña de al menos 8 caracteres."
        : (error.status === 400 || error.status === 401)
          ? "No pudimos validar esos datos. Revisa el correo o crea una cuenta nueva."
          : "No pudimos conectar con Trama. Inténtalo de nuevo en un momento.");
    } finally {
      authSubmit.disabled = false;
    }
  }

  async function handleGoalSubmit(event) {
    event.preventDefault();
    const form = event.currentTarget;
    const formData = new FormData(form);
    const domain = String(formData.get("domain") || "").trim();
    const purpose = String(formData.get("purpose") || "").trim();
    const level = String(formData.get("level") || "inicial");
    const minutes = Number(formData.get("minutes_per_week"));
    if (!domain || domain.length > 100 || !purpose || purpose.length > 240 || !Number.isInteger(minutes) || minutes < 10 || minutes > 10080) {
      showNotice("Revisa el tema, propósito y tiempo disponible.", true);
      return;
    }
    const button = form.querySelector("button[type=submit]");
    button.disabled = true;
    try {
      const goal = await api("/api/collections/learning_goals/records", {
        method: "POST",
        body: {
          user: account.id,
          domain,
          level,
          purpose,
          minutes_per_week: minutes,
          status: "active",
          progress: 0,
          path: starterPath(domain, purpose, level),
        },
      });
      goals = [goal, ...goals];
      renderWorkspace();
      renderGoal(goal);
      form.reset();
      byId("minutes").value = "90";
      showNotice("Tu nueva ruta quedó guardada en tu espacio privado.");
    } catch (error) {
      showNotice("No pudimos guardar la ruta. Vuelve a intentarlo.", true);
    } finally {
      button.disabled = false;
    }
  }

  async function handleAttemptSubmit(event) {
    if (!event.target.matches(".attempt-form")) return;
    event.preventDefault();
    const form = event.target;
    const button = form.querySelector("button[type=submit]");
    button.disabled = true;
    try {
      const result = await api(`/api/trama/goals/${encodeURIComponent(form.dataset.goalId)}/attempts`, {
        method: "POST",
        body: { index: Number(form.dataset.index), answer: new FormData(form).get("answer") },
      });
      await ensureInitialSession();
      const goal = goals.find((item) => item.id === form.dataset.goalId);
      if (goal) renderGoal(goal);
      showNotice(`${result.assessment.correct ? "Buen comienzo" : "Respuesta guardada"}. Avance de esta ruta: ${(Number(result.progress) * 100).toFixed(0)}%. La evaluación es una demostración heurística.`);
    } catch (error) {
      showNotice(error.status === 404 ? "No encontramos esa ruta en tu cuenta." : "No pudimos guardar la respuesta. Inténtalo de nuevo.", true);
    } finally {
      button.disabled = false;
    }
  }

  async function restoreAuth() {
    if (!token) return false;
    try {
      const refreshed = await api("/api/collections/users/auth-refresh", { method: "POST" });
      saveAuth(refreshed);
      return true;
    } catch (error) {
      if (error.status === 0 || !error.status) {
        showNotice("No pudimos conectar con Trama. Revisa tu conexión e inténtalo de nuevo.", true);
        return false;
      }
      clearAuth();
      return false;
    }
  }

  authForm.addEventListener("submit", handleAuthSubmit);
  authSwitch.addEventListener("click", () => {
    authMode = authMode === "login" ? "register" : "login";
    renderAuthMode();
  });
  byId("goal-form").addEventListener("submit", handleGoalSubmit);
  byId("goal-list").addEventListener("click", (event) => {
    const button = event.target.closest("[data-open-goal]");
    if (!button) return;
    const goal = goals.find((item) => item.id === button.dataset.openGoal);
    if (goal) renderGoal(goal);
  });
  byId("goal-detail").addEventListener("submit", (event) => {
    if (event.target.matches(".attempt-form")) handleAttemptSubmit(event);
  });
  byId("logout-button").addEventListener("click", () => {
    clearAuth();
    goals = [];
    workspace.hidden = true;
    byId("account-controls").hidden = true;
    authScreen.hidden = false;
    showNotice("Cerraste sesión en esta pestaña.");
  });

  (async () => {
    if (await restoreAuth()) {
      try {
        await ensureInitialSession();
      } catch {
        clearAuth();
        workspace.hidden = true;
        authScreen.hidden = false;
        showNotice("No pudimos cargar tu espacio. Inicia sesión de nuevo.", true);
      }
    }
    renderAuthMode();
  })();
})();
