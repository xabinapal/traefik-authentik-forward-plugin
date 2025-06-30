import path from "path";

import { chromium } from "@playwright/test";

import * as dockerCompose from "docker-compose";

export default async function globalSetup() {
  console.log("Starting docker compose sandbox services...");

  await dockerCompose.upAll({
    cwd: path.join(__dirname, "../sandbox"),
    commandOptions: ["--wait"],
  });

  console.log("Started Docker compose sandbox services");

  console.log("Waiting for docker compose sandbox services to be ready...");

  const browser = await chromium.launch();
  const page = await browser.newPage();

  let traefikStatusCode = 0;
  while (traefikStatusCode !== 200) {
    const traefik = await page.goto("http://traefik.localhost:8080/ping/");
    traefikStatusCode = traefik?.status() || 0;

    if (traefikStatusCode !== 200) {
      console.log("Traefik not ready yet...");
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
  }

  let authentikStatusCode = 0;
  while (authentikStatusCode !== 200) {
    const authentik = await page.goto(
      "http://authentik.localhost:9000/-/health/ready/",
    );
    authentikStatusCode = authentik?.status() || 0;

    if (authentikStatusCode !== 200) {
      console.log("Authentik not ready yet...");
      await new Promise((resolve) => setTimeout(resolve, 1000));
    }
  }

  await browser.close();

  console.log("Docker compose sandbox services are ready");
}
