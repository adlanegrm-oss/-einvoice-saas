import concurrent.futures
import json
import time
import requests

BASE_URL = "http://127.0.0.1:8080"
ADMIN_CREDS = {"email": "admin@example.com", "password": "PasswordAdmin123!"}
CLIENT_CREDS = {"email": "client@example.com", "password": "PasswordClient123!"}

class SaaSInspector:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.tokens = {}

    def log(self, section: str, message: str, status: str = "INFO"):
        symbols = {"PASS": "[+]", "FAIL": "[-]", "WARN": "[!]", "INFO": "[*]"}
        print(f"{symbols.get(status, '[*]')} [{section}] {message}")

    def authenticate(self, role: str, creds: dict):
        url = f"{self.base_url}/api/v1/auth/login"
        try:
            r = self.session.post(url, json=creds, timeout=5)
            if r.status_code == 200:
                data = r.json()
                self.tokens[role] = data.get("token") or data.get("access_token")
                self.log("AUTH", f"Authentification {role} réussie", "PASS")
            else:
                self.log("AUTH", f"Échec auth {role} (HTTP {r.status_code})", "WARN")
        except Exception as e:
            self.log("AUTH", f"Impossible de joindre le serveur: {e}", "FAIL")

    # --- TESTS DE SÉCURITÉ ---

    def test_unauthenticated_access(self):
        """Vérifie le blocage 401 sur les routes privées."""
        protected_endpoints = [
            ("GET", "/api/v1/auth/me"),
            ("GET", "/api/v1/invoices"),
            ("GET", "/api/v1/invoices/list"),
            ("POST", "/api/v1/invoices/deposit"),
            ("GET", "/api/v1/reports/daily?date=2026-10-01"),
        ]
        passed = True
        for method, ep in protected_endpoints:
            r = self.session.request(method, f"{self.base_url}{ep}")
            if r.status_code != 401:
                self.log("AUTH-BYPASS", f"Accès non autorisé possible sur {ep} (HTTP {r.status_code})", "FAIL")
                passed = False
        if passed:
            self.log("AUTH-BYPASS", "Toutes les routes protégées retournent HTTP 401 sans jeton", "PASS")

    def test_privilege_escalation(self):
        """Vérifie qu'un rôle CLIENT ne peut pas déclencher les jobs ADMIN."""
        client_token = self.tokens.get("CLIENT")
        if not client_token:
            self.log("RBAC", "Test ignoré: jeton CLIENT indisponible", "WARN")
            return

        headers = {"Authorization": f"Bearer {client_token}"}
        admin_routes = [
            ("POST", "/api/v1/jobs/daily-report"),
            ("POST", "/api/v1/cron/archive-job"),
        ]
        passed = True
        for method, ep in admin_routes:
            r = self.session.request(method, f"{self.base_url}{ep}", headers=headers)
            if r.status_code != 403:
                self.log("RBAC", f"Élévation possible: {ep} accessible par CLIENT (HTTP {r.status_code})", "FAIL")
                passed = False
        if passed:
            self.log("RBAC", "Isolation des privilèges validée (HTTP 403 renvoyé au CLIENT)", "PASS")

    def test_xxe_injection(self):
        """Injecte des entités externes XML sur l'endpoint de validation."""
        token = self.tokens.get("CLIENT") or self.tokens.get("ADMIN")
        headers = {"Authorization": f"Bearer {token}"} if token else {}
        
        xxe_payload = """<?xml version="1.0" encoding="UTF-8"?>
        <!DOCTYPE foo [ <!ENTITY xxe SYSTEM "file:///etc/passwd"> ]>
        <Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2">
            <ID>&xxe;</ID>
        </Invoice>"""

        r = self.session.post(
            f"{self.base_url}/api/v1/validate",
            data=xxe_payload,
            headers=headers,
            timeout=5
        )
        if "root:" in r.text:
            self.log("XXE", "Vulnérabilité critique: lecture de fichier /etc/passwd réussie !", "FAIL")
        elif r.status_code in [400, 422]:
            self.log("XXE", f"Injection XML/DOCTYPE correctement neutralisée (HTTP {r.status_code})", "PASS")
        else:
            self.log("XXE", f"Comportement inattendu lors de l'injection (HTTP {r.status_code})", "WARN")

    def test_path_traversal(self):
        """Teste la fuite de chemin sur le téléchargement et l'export."""
        token = self.tokens.get("CLIENT") or self.tokens.get("ADMIN")
        headers = {"Authorization": f"Bearer {token}"} if token else {}

        payloads = [
            "../../../../etc/passwd",
            "..%2F..%2F..%2F..%2Fetc%2Fpasswd",
            "....//....//....//etc/passwd",
        ]
        passed = True
        for p in payloads:
            # Test sur /download
            r1 = self.session.get(f"{self.base_url}/api/v1/invoices/download?file={p}", headers=headers)
            # Test sur /export
            r2 = self.session.get(f"{self.base_url}/api/v1/invoices/export?id={p}", headers=headers)

            for target, resp in [("download", r1), ("export", r2)]:
                if "root:x:" in resp.text:
                    self.log("PATH-TRAVERSAL", f"Fuite de fichier via {target} avec le chemin: {p}", "FAIL")
                    passed = False
                elif resp.status_code not in [400, 404]:
                    self.log("PATH-TRAVERSAL", f"Code anormal ({resp.status_code}) sur {target} avec: {p}", "WARN")
        if passed:
            self.log("PATH-TRAVERSAL", "Aucune traversée de répertoire détectée", "PASS")

    def test_rate_limiting(self):
        """Contrôle le déclenchement de HTTP 429 lors d'un burst de requêtes."""
        url = f"{self.base_url}/api/v1/auth/login"
        hit_limit = False
        self.log("RATE-LIMIT", "Lancement d'une rafale de 45 tentatives de login...", "INFO")
        for i in range(45):
            r = self.session.post(url, json={"email": "audit@probe.com", "password": "wrong"}, timeout=3)
            if r.status_code == 429:
                hit_limit = True
                self.log("RATE-LIMIT", f"Protection activée avec succès après {i+1} requêtes (HTTP 429)", "PASS")
                break
            time.sleep(0.02)
        if not hit_limit:
            self.log("RATE-LIMIT", "Aucun blocage HTTP 429 observé après 45 requêtes rapprochées", "WARN")

    # --- TESTS DE PERFORMANCES ---

    def test_performance_benchmark(self, total_requests=100, concurrency=10):
        """Mesure la latence sur la sonde /health et le traitement sous concurrence."""
        url = f"{self.base_url}/health"
        latencies = []
        errors = 0

        def send_request():
            start = time.perf_counter()
            try:
                r = requests.get(url, timeout=5)
                duration = time.perf_counter() - start
                return (r.status_code == 200, duration)
            except Exception:
                return (False, 0)

        self.log("PERF", f"Benchmark en cours: {total_requests} requêtes (concurrence: {concurrency})...", "INFO")
        bench_start = time.perf_counter()

        with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
            futures = [executor.submit(send_request) for _ in range(total_requests)]
            for future in concurrent.futures.as_completed(futures):
                success, duration = future.result()
                if success:
                    latencies.append(duration)
                else:
                    errors += 1

        total_time = time.perf_counter() - bench_start
        rps = len(latencies) / total_time if total_time > 0 else 0

        if latencies:
            latencies.sort()
            avg_ms = (sum(latencies) / len(latencies)) * 1000
            p95_ms = latencies[int(len(latencies) * 0.95)] * 1000
            min_ms = latencies[0] * 1000
            max_ms = latencies[-1] * 1000

            print("\n" + "="*45 + " RÉSULTATS DE PERFORMANCE " + "="*45)
            print(f"  Débit moyen        : {rps:.2f} requêtes/sec")
            print(f"  Temps d'exécution  : {total_time:.2f} s")
            print(f"  Succès             : {len(latencies)}/{total_requests} ({errors} échecs)")
            print(f"  Latence Min        : {min_ms:.2f} ms")
            print(f"  Latence Moyenne    : {avg_ms:.2f} ms")
            print(f"  Latence p95        : {p95_ms:.2f} ms")
            print(f"  Latence Max        : {max_ms:.2f} ms")
            print("="*116 + "\n")
        else:
            self.log("PERF", "Toutes les requêtes de benchmark ont échoué", "FAIL")

    def run_all(self):
        print("\n" + "#"*30 + f" AUDIT DE SÉCURITÉ & PERF : {self.base_url} " + "#"*30 + "\n")
        self.authenticate("ADMIN", ADMIN_CREDS)
        self.authenticate("CLIENT", CLIENT_CREDS)
        print("-" * 80)
        self.test_unauthenticated_access()
        self.test_privilege_escalation()
        self.test_xxe_injection()
        self.test_path_traversal()
        self.test_rate_limiting()
        print("-" * 80)
        self.test_performance_benchmark(total_requests=100, concurrency=10)


if __name__ == "__main__":
    inspector = SaaSInspector(BASE_URL)
    inspector.run_all()