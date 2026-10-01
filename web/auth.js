// Utilitaires d'authentification partagés par les pages du portail.
// NB : le contrôle d'accès réel est fait par l'API (jeton signé). Ce fichier
// ne sert qu'à rediriger vers la connexion et à joindre le jeton aux requêtes.
(function () {
    function token() { return sessionStorage.getItem('authToken') || ''; }
    function role() { return sessionStorage.getItem('userRole') || ''; }

    function logout() {
        sessionStorage.clear();
        window.location.replace('/login.html');
    }

    // Redirige vers la connexion si l'utilisateur n'a pas l'un des rôles requis.
    function require(roles) {
        if (!token() || roles.indexOf(role()) === -1) {
            logout();
            return false;
        }
        return true;
    }

    // fetch() avec le jeton ; une réponse 401 (jeton expiré) ramène à la connexion.
    async function api(path, opts) {
        opts = opts || {};
        opts.headers = Object.assign({}, opts.headers, { Authorization: 'Bearer ' + token() });
        const res = await fetch(path, opts);
        if (res.status === 401) {
            logout();
            throw new Error('Session expirée');
        }
        return res;
    }

    window.Auth = { token: token, role: role, logout: logout, require: require, api: api };

    // Échappement HTML pour toute donnée insérée via innerHTML (anti-XSS).
    window.escapeHtml = function (value) {
        return String(value == null ? '' : value)
            .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    };
})();
