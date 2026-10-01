// Dictionnaire exhaustif des formulaires par protocole technique
const PROTOCOL_CONFIG_SCHEMAS = {
    AS2: `
        <div class="grid-2col">
            <div class="form-group">
                <label>AS2 Identifier Local (AS2-From) :</label>
                <input type="text" class="form-control" placeholder="ex: TOTAL_ENERGIES_PROD_AS2" />
            </div>
            <div class="form-group">
                <label>AS2 Identifier Distant (AS2-To) :</label>
                <input type="text" class="form-control" placeholder="ex: CARREFOUR_HUB_AS2" />
            </div>
        </div>
        <div class="form-group">
            <label>URL Endpoint AS2 (HTTP/S) :</label>
            <input type="url" class="form-control" placeholder="https://as2.partenaire.com:8443/as2/receive" />
        </div>
        <div class="grid-2col">
            <div class="form-group">
                <label>Certificat Chiffrement X.509 (.cer/.pem) :</label>
                <input type="file" class="form-control" accept=".cer,.crt,.pem" />
            </div>
            <div class="form-group">
                <label>Accusé de réception MDN :</label>
                <select class="form-control">
                    <option value="SYNC_SIGNED">Synchrone Signé (Recommandé)</option>
                    <option value="ASYNC_SIGNED">Asynchrone Signé (Grand Volume)</option>
                    <option value="NONE">Non signé</option>
                </select>
            </div>
        </div>
    `,

    X400: `
        <div class="grid-3col">
            <div class="form-group">
                <label>Country Code (C) :</label>
                <input type="text" class="form-control" value="FR" />
            </div>
            <div class="form-group">
                <label>ADMD (Opérateur Public) :</label>
                <input type="text" class="form-control" placeholder="ex: ATLAS, TELECOM" />
            </div>
            <div class="form-group">
                <label>PRMD (Domaine Privé) :</label>
                <input type="text" class="form-control" placeholder="ex: TOTAL, GEIS" />
            </div>
        </div>
        <div class="grid-3col">
            <div class="form-group">
                <label>Organization (O) :</label>
                <input type="text" class="form-control" placeholder="ex: TOTAL-MARKETING" />
            </div>
            <div class="form-group">
                <label>Organizational Unit (OU) :</label>
                <input type="text" class="form-control" placeholder="ex: EDI-FACTURATION" />
            </div>
            <div class="form-group">
                <label>Mailbox / User (CN/S) :</label>
                <input type="text" class="form-control" placeholder="ex: BAL_PROD_01" />
            </div>
        </div>
        <div class="form-group">
            <label>Code Boîte Réseau VAN (GXS / OpenText / EDICOM) :</label>
            <input type="text" class="form-control" placeholder="ex: VAN_BOX_FR_99872" />
        </div>
    `,

    PEPPOL_AS4: `
        <div class="grid-2col">
            <div class="form-group">
                <label>Peppol Participant ID (Scheme + ID) :</label>
                <input type="text" class="form-control" placeholder="ex: 0002:40320501800012 ou 9957:FR..." />
            </div>
            <div class="form-group">
                <label>Point d'Accès Certifié (SMP) :</label>
                <input type="text" class="form-control" placeholder="ex: http://smp.peppol.org/..." />
            </div>
        </div>
        <div class="grid-2col">
            <div class="form-group">
                <label>Process ID (Profil Métier) :</label>
                <input type="text" class="form-control" value="urn:fdc:peppol.eu:2017:poacc:billing:01:1.0" />
            </div>
            <div class="form-group">
                <label>Document Identifier :</label>
                <input type="text" class="form-control" value="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2" />
            </div>
        </div>
    `,

    OFTP2: `
        <div class="grid-2col">
            <div class="form-group">
                <label>SSID Émetteur Local :</label>
                <input type="text" class="form-control" placeholder="ex: O0013000001234TOTAL" />
            </div>
            <div class="form-group">
                <label>SSID Destinataire Distant :</label>
                <input type="text" class="form-control" placeholder="ex: O0013000005678PARTNER" />
            </div>
        </div>
        <div class="grid-2col">
            <div class="form-group">
                <label>Hôte OFTP2 & Port :</label>
                <input type="text" class="form-control" placeholder="oftp2.partenaire.com:3305" />
            </div>
            <div class="form-group">
                <label>Passphrase Connexion :</label>
                <input type="password" class="form-control" placeholder="••••••••••••" />
            </div>
        </div>
    `,

    SFTP: `
        <div class="grid-3col">
            <div class="form-group">
                <label>Hôte SFTP :</label>
                <input type="text" class="form-control" placeholder="sftp.partenaire.fr" />
            </div>
            <div class="form-group">
                <label>Port :</label>
                <input type="number" class="form-control" value="22" />
            </div>
            <div class="form-group">
                <label>Utilisateur / Compte :</label>
                <input type="text" class="form-control" placeholder="edi_user" />
            </div>
        </div>
        <div class="grid-2col">
            <div class="form-group">
                <label>Répertoire Dépose (IN) :</label>
                <input type="text" class="form-control" value="/data/inbound/" />
            </div>
            <div class="form-group">
                <label>Répertoire Récupération (OUT) :</label>
                <input type="text" class="form-control" value="/data/outbound/" />
            </div>
        </div>
    `,

    SAP_ALE: `
        <div class="grid-3col">
            <div class="form-group">
                <label>SAP Application Host (IP/DNS) :</label>
                <input type="text" class="form-control" placeholder="sap-app.internal.domain" />
            </div>
            <div class="form-group">
                <label>System Number (Sysnr) :</label>
                <input type="text" class="form-control" placeholder="00" />
            </div>
            <div class="form-group">
                <label>Client / Mandant SAP :</label>
                <input type="text" class="form-control" placeholder="100" />
            </div>
        </div>
        <div class="grid-2col">
            <div class="form-group">
                <label>RFC Program ID :</label>
                <input type="text" class="form-control" placeholder="EINVOICE_RFC_SERVER" />
            </div>
            <div class="form-group">
                <label>Gateway Host / Gateway Service :</label>
                <input type="text" class="form-control" placeholder="sapgw00 / sapdp00" />
            </div>
        </div>
    `,

    REST_API: `
        <div class="form-group">
            <label>URL Endpoint de l'API / Webhook :</label>
            <input type="url" class="form-control" placeholder="https://api.pdp-agree.fr/v1/invoices/dispatch" />
        </div>
        <div class="grid-2col">
            <div class="form-group">
                <label>Mode d'Authentification :</label>
                <select class="form-control">
                    <option value="BEARER">JWT Bearer Token / OAuth2</option>
                    <option value="MTLS">Mutual TLS (mTLS Certificat Client)</option>
                    <option value="API_KEY">X-API-Key En-tête Custom</option>
                </select>
            </div>
            <div class="form-group">
                <label>Jeton / Secret / Clé d'API :</label>
                <input type="password" class="form-control" placeholder="Secret Token ou Clé d'accès" />
            </div>
        </div>
    `
};

// Rendu dynamique selon le rôle et le protocole choisi
function renderConnectorFields(target) {
    const selectElem = document.getElementById(`${target}_protocol_select`);
    const container = document.getElementById(`${target}_dynamic_fields`);
    const selectedProtocol = selectElem.value;

    if (PROTOCOL_CONFIG_SCHEMAS[selectedProtocol]) {
        container.innerHTML = PROTOCOL_CONFIG_SCHEMAS[selectedProtocol];
    } else {
        container.innerHTML = "<p class='text-muted'>Aucun paramètre supplémentaire requis.</p>";
    }
}

// Initialisation au chargement de l'interface
document.addEventListener("DOMContentLoaded", () => {
    renderConnectorFields("do");
    renderConnectorFields("partner");
});

// Simulation de validation technique
function testFullConnectivity() {
    alert("Vérification en cours :\n- Résolution DNS des endpoints\n- Validation de la chaîne de confiance X.509\n- Handshake TLS / Accès répertoires SFTP\n\nRésultat : LIAISON ÉTABLIE AVEC SUCCÈS !");
}

function saveFlowImplementation() {
    const doClient = document.getElementById("flow_client_do").options[document.getElementById("flow_client_do").selectedIndex].text;
    const flowType = document.getElementById("flow_type").options[document.getElementById("flow_type").selectedIndex].text;
    const doProto = document.getElementById("do_protocol_select").options[document.getElementById("do_protocol_select").selectedIndex].text;
    const partnerProto = document.getElementById("partner_protocol_select").options[document.getElementById("partner_protocol_select").selectedIndex].text;
    const mapping = document.getElementById("flow_mapping_engine").options[document.getElementById("flow_mapping_engine").selectedIndex].text;

    const tableBody = document.getElementById("flows_table_body");
    const newRow = document.createElement("tr");

    newRow.innerHTML = `
        <td><strong>${doClient}</strong></td>
        <td>${flowType}</td>
        <td><span class="tag-tech tag-sap">${doProto}</span></td>
        <td><span class="tag-tech tag-peppol">${partnerProto}</span></td>
        <td>${mapping}</td>
        <td><span class="badge-active">ACTIF</span></td>
        <td><button class="btn-delete" onclick="this.closest('tr').remove()">Supprimer</button></td>
    `;

    tableBody.prepend(newRow);
    alert("Nouveau flux technique enregistré et déployé avec succès !");
}