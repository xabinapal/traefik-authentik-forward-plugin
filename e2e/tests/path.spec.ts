import { expect } from "@playwright/test";
import { StatusCodes } from "http-status-codes";

import { test } from "../fixtures";

test.describe("restricted", () => {
  [
    {
      path: "/outpost.goauthentik.io/start",
      status: StatusCodes.MOVED_TEMPORARILY,
    },
    {
      path: "/outpost.goauthentik.io/sign_out",
      status: StatusCodes.MOVED_TEMPORARILY,
    },
    {
      path: "/outpost.goauthentik.io/callback",
      status: StatusCodes.BAD_REQUEST,
    },
  ].forEach(({ path, status }) => {
    test(`should return ${status} with ${path}`, async ({ request }) => {
      const response = await request.get(`http://whoami.localhost/${path}`, {
        maxRedirects: 0,
      });
      expect(response.status()).toBe(status);
    });
  });
});

test.describe("allowed", () => {
  [
    "/outpost.goauthentik.io",
    "/outpost.goauthentik.io/auth/nginx",
    "/outpost.goauthentik.io/auth/traefik",
    "/outpost.goauthentik.io/auth/caddy",
    "/outpost.goauthentik.io/auth/envoy",
  ].forEach((path) => {
    test(`should return 404 with ${path}`, async ({ request }) => {
      const response = await request.get(`http://whoami.localhost/${path}`);
      expect(response.status()).toBe(StatusCodes.NOT_FOUND);
    });
  });
});
