import { expect } from "@playwright/test";
import { StatusCodes } from "http-status-codes";

import { test } from "../fixtures";

test("should return unauthorized with /deny", async ({ request }) => {
  const response = await request.get("http://whoami.localhost/deny");
  expect(response.status()).toBe(StatusCodes.UNAUTHORIZED);
});

test("should redirect to start flow with /login", async ({ request }) => {
  const response = await request.get("http://whoami.localhost/login", {
    maxRedirects: 0,
  });
  expect(response.status()).toBe(StatusCodes.MOVED_TEMPORARILY);

  const location = response.headers()["location"];
  expect(location).toBeDefined();

  const locationUrl = new URL(location);
  expect(locationUrl.protocol).toBe("http:");
  expect(locationUrl.hostname).toBe("whoami.localhost");
  expect(locationUrl.pathname).toBe("/outpost.goauthentik.io/start");
  expect(locationUrl.searchParams.get("rd")).toBe(
    "http://whoami.localhost/login",
  );
});
