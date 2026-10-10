import os
from reportlab.lib.pagesizes import A4
from reportlab.pdfgen import canvas

os.makedirs("test_invoices", exist_ok=True)

invoices = [
    {"num": "FACT-2026-002", "client": "Alpha Tech SARL", "amount": "4 500,00 €"},
    {"num": "FACT-2026-003", "client": "Beta Solutions Corp", "amount": "12 150,00 €"},
    {"num": "FACT-2026-004", "client": "Gamma Logistics", "amount": "2 800,00 €"},
    {"num": "FACT-2026-005", "client": "Delta Digital Agency", "amount": "6 320,00 €"}
]

for inv in invoices:
    filename = f"test_invoices/{inv['num']}.pdf"
    c = canvas.Canvas(filename, pagesize=A4)
    width, height = A4

    # En-tête
    c.setFont("Helvetica-Bold", 16)
    c.drawString(50, height - 50, "APEX CLOUD SOLUTIONS SAS")
    c.setFont("Helvetica", 10)
    c.drawString(50, height - 65, "128 Boulevard Haussmann, 75008 Paris")
    c.drawString(50, height - 80, "SIRET: 849 201 932 - TVA: FR 82 849201932")

    # Informations facture
    c.setFont("Helvetica-Bold", 12)
    c.drawString(50, height - 130, f"Facture N° : {inv['num']}")
    c.setFont("Helvetica", 10)
    c.drawString(50, height - 150, f"Client : {inv['client']}")
    c.drawString(50, height - 165, "Date : 10/10/2026")

    # Tableau / Lignes
    c.line(50, height - 190, width - 50, height - 190)
    c.drawString(50, height - 210, "Description de la prestation")
    c.drawString(400, height - 210, "Total TTC")
    c.line(50, height - 220, width - 50, height - 220)

    c.drawString(50, height - 240, "Prestation de services informatiques & Cloud")
    c.drawString(400, height - 240, inv['amount'])

    # Pied de page
    c.line(50, 100, width - 50, 100)
    c.setFont("Helvetica", 8)
    c.drawString(50, 85, "Document test pour audit EN 16931 / Factur-X")
    
    c.save()
    print(f"Généré : {filename}")

print("Génération du lot terminée avec succès dans le dossier 'test_invoices/' !")
