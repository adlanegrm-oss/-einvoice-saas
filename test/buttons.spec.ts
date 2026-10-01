import { test, expect } from '@playwright/test';

test('Test des boutons et de la navigation du portail B2B', async ({ page }) => {
  // 1. Accès à la page d'accueil
  await page.goto('http://localhost:8080/');
  await expect(page).toHaveTitle(/eInvoice/i);

  // 2. Navigation dans les menus principaux
  await page.getByRole('link', { name: 'Commandes Clients' }).click();
  await page.getByRole('link', { name: 'Factures Clients' }).click();
  await page.getByRole('link', { name: 'Aides & Support' }).click();

  // 3. Actions et raccourcis
  await page.getByRole('link', { name: 'Déposer une Facture PDF /' }).click();
  await page.getByRole('link', { name: 'Consulter les Factures' }).click();
  await page.getByRole('link', { name: 'Suivi des Traitements' }).click();
});