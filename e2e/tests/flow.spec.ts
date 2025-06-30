import { Response, expect } from "@playwright/test";
import { StatusCodes } from "http-status-codes";

import { test } from "../fixtures";

test.describe("authentication", () => {
  test.describe.configure({ mode: "serial" });

  test("should return upstream after login", async ({
    sharedContextPage: page,
  }) => {
    // go to login page
    await page.goto("http://whoami.localhost/login");

    // set redirect wait
    const responsePromise = page.waitForResponse(
      "http://whoami.localhost/login",
      { timeout: 30000 },
    );

    // enter authentik username
    await page.waitForSelector(
      "ak-stage-identification ak-form-element input[name=uidField]",
    );
    await page.fill(
      "ak-stage-identification ak-form-element input[name=uidField]",
      "akadmin",
    );
    page.click("ak-stage-identification button[type=submit]");

    // enter authentik password
    await page.waitForSelector(
      "ak-stage-password ak-form-element input[name=password]",
    );
    page.fill(
      "ak-stage-password ak-form-element input[name=password]",
      "authentik",
    );
    page.click("ak-stage-password button[type=submit]");

    // wait for redirect
    const response = await responsePromise;
    await response.finished();

    // check for upstream
    expect(response.status()).toBe(StatusCodes.OK);
    expect(await page.content()).toContain("X-Authentik-Username: akadmin");
  });

  test("should stay logged on other pages", async ({
    sharedContextPage: page,
  }) => {
    // go to main page
    const response = (await page.goto("http://whoami.localhost")) as Response;

    // check for upstream
    expect(response.status()).toBe(StatusCodes.OK);
    expect(await page.content()).toContain("X-Authentik-Username: akadmin");
  });

  test("should return unauthorized after logout", async ({
    sharedContextPage: page,
  }) => {
    // go to logout page
    await page.goto("http://whoami.localhost/outpost.goauthentik.io/sign_out");

    // wait for redirect
    await page.waitForURL("http://authentik.localhost:9000/**");

    // go to main page
    const response = (await page.goto("http://whoami.localhost")) as Response;

    // check for upstream
    expect(response.status()).toBe(StatusCodes.UNAUTHORIZED);
  });
});
