import path from "path";

import * as dockerCompose from "docker-compose";

export default async function globalTeardown() {
  console.log("Stopping docker compose sandbox services...");

  await dockerCompose.downAll({
    cwd: path.join(__dirname, "../sandbox"),
  });

  console.log("Stopped docker compose sandbox services");
}
