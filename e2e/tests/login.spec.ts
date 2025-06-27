import { expect } from "@playwright/test";
import { StatusCodes } from "http-status-codes";

import { test } from "../fixtures";

test.describe("flow", () => {
  test.describe.configure({ mode: "serial" });

  let context: any;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
  });

  test.afterAll(async () => {
    await context?.close();
  });

  test("should return upstream after login", async () => {
    const page = await context.newPage();
    await page.goto("http://whoami.localhost/login");
    const loggedResponsePromise = page.waitForResponse(
      "http://whoami.localhost/login",
      { timeout: 500000 },
    );

    // enter authentik username
    page.fill("input[name=uidField]", "akadmin");
    page.click("button[type=submit]");

    // wait for password input
    await page.waitForSelector("input[name=password]:visible");
    await page.waitForTimeout(1000);

    // enter authentik password
    page.fill("input[name=password]:visible", "authentik");
    await page.click("button[type=submit]");

    // check for redirect
    const loggedResponse = await loggedResponsePromise;
    await page.waitForEvent("requestfinished");

    // check for upstream
    expect(loggedResponse.status()).toBe(StatusCodes.OK);
    expect(await page.content()).toContain("X-Authentik-Username: akadmin");
  });

  test("should stay logged on other pages", async () => {
    const page = await context.newPage();
    const response = await page.goto("http://whoami.localhost");

    // check for upstream
    expect(response?.status()).toBe(StatusCodes.OK);
    expect(await page.content()).toContain("X-Authentik-Username: akadmin");
  });

  test("should return unauthorized after logout", async () => {
    const page = await context.newPage();
    await page.goto("http://whoami.localhost/outpost.goauthentik.io/sign_out");
    await page.waitForURL("http://authentik.localhost:9000/**");

    const response = await page.goto("http://whoami.localhost");

    // check for upstream
    expect(response?.status()).toBe(StatusCodes.UNAUTHORIZED);
  });
});
