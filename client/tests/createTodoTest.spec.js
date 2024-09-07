import { test, expect } from "@playwright/test";

test("test", async ({ page }) => {
  await page.goto("http://localhost:5173/");
  await page.getByRole("button", { name: "ADD TODO" }).click();
  await page.getByPlaceholder("What is your todo?").click();
  await page.getByPlaceholder("What is your todo?").fill("playwrighttest");
  await page.getByPlaceholder("Tell me more...").click();
  await page.getByPlaceholder("Tell me more...").fill("test");
  await page.getByPlaceholder("Enter category").click();
  await page.getByPlaceholder("Enter category").fill("testcategory");
  await page.getByRole("button", { name: "Save" }).click();

  // const elements = page.locator("text=playwrighttesttesttestcategory✎☒✘");
  // await expect(elements).toHaveCount(1); // Adjust the count as needed
  // await expect(elements.first()).toBeVisible();

  await page.waitForSelector("text=playwright");
  await expect(page.locator("text=playwright")).toBeVisible();

  //
  // // Assert that the text exists
  // await expect(
  //   page.getByText("playwrighttesttesttestcategory✎☒✘"),
  // ).toBeVisible();
});
