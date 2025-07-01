import { expect } from "@playwright/test";
import { StatusCodes } from "http-status-codes";

import { test } from "../fixtures";

test("should return ok with /allow", async ({ request }) => {
  const response = await request.get("http://whoami.localhost/allow");
  expect(response.status()).toBe(StatusCodes.OK);

  const body = await response.text();
  expect(body).toContain("X-Forwarded-Host: whoami.localhost");
  expect(body).not.toContain("X-Authentik-User");
});

test("should return unauthorized with /deny", async ({ request }) => {
  const response = await request.get("http://whoami.localhost/deny");
  expect(response.status()).toBe(StatusCodes.UNAUTHORIZED);
});

test("should redirect to start flow with /login", async ({ request }) => {
  const response = await request.get("http://whoami.localhost/login");
  expect(response.status()).toBe(StatusCodes.MOVED_TEMPORARILY);

  const location = response.headers()["location"] as string;
  expect(location).toBeDefined();

  const locationUrl = new URL(location);
  expect(locationUrl.protocol).toBe("http:");
  expect(locationUrl.hostname).toBe("whoami.localhost");
  expect(locationUrl.pathname).toBe("/outpost.goauthentik.io/start");
  expect(locationUrl.searchParams.get("rd")).toBe(
    "http://whoami.localhost/login",
  );
});

test("should redirect to start flow with /sign_out", async ({ request }) => {
  const response = await request.get(
    "http://whoami.localhost/outpost.goauthentik.io/sign_out",
  );

  expect(response.status()).toBe(StatusCodes.MOVED_TEMPORARILY);

  const location = response.headers()["location"] as string;
  expect(location).toBeDefined();

  const locationUrl = new URL(location);
  expect(locationUrl.protocol).toBe("http:");
  expect(locationUrl.hostname).toBe("whoami.localhost");
  expect(locationUrl.pathname).toBe("/outpost.goauthentik.io/start");
  expect(locationUrl.searchParams.get("rd")).toBe(
    "http://whoami.localhost/outpost.goauthentik.io/sign_out",
  );
});

test("should return bad request with invalid /callback", async ({
  request,
}) => {
  const response = await request.get(
    "http://whoami.localhost/outpost.goauthentik.io/callback",
  );

  expect(response.status()).toBe(StatusCodes.BAD_REQUEST);
});
