import { test, expect } from "@playwright/test";

test("test", async ({ page }) => {
	await page.goto("http://localhost:5173/");

	// Open the 'Add Todo' form
	await page.getByRole("button", { name: "ADD TODO" }).click();

	// Fill out the form
	await page.getByPlaceholder("What is your todo?").fill("playwrighttest");
	await page.getByPlaceholder("Tell me more...").fill("test");
	await page.getByPlaceholder("Enter category").fill("testcategory");

	// Submit the form
	await page.getByRole("button", { name: "Save" }).click();

	// Wait for the newly added todo to appear in the list
	await page.waitForSelector("text=playwrighttest");

	// More specific locator for the exact task we added
	const todoItem = page.locator("text=playwrighttest").first();

	// Ensure the todo is visible
	await expect(todoItem).toBeVisible();
});

/* REMEMBER TO DO CLEAN UP NEXT TIME */
