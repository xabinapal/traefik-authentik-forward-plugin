import {
  test as base,
  request as baseRequest,
  APIRequestContext,
  BrowserContext,
  Page,
} from "@playwright/test";

interface TestFixtures {
  request: APIRequestContext;
  sharedContextPage: Page;
}

interface WorkerFixtures {
  sharedContext: BrowserContext;
}

export const test = base.extend<TestFixtures, WorkerFixtures>({
  request: async ({}, use) => {
    const context = await baseRequest.newContext();
    const originalFetch = context.fetch.bind(context);

    context.fetch = async (url, options = {}) => {
      const originalUrl: URL =
        typeof url === "string" ? new URL(url) : new URL(url.url());

      const rewrittenUrl = new URL(originalUrl.href);
      rewrittenUrl.host = "localhost";

      options.headers = {
        ...options.headers,
        Host: originalUrl.host,
      };

      options.maxRedirects = 0;

      return originalFetch(rewrittenUrl.href, options);
    };

    await use(context);
    await context.dispose();
  },

  sharedContext: [
    async ({ browser }, use) => {
      const context = await browser.newContext();
      await use(context);
      await context.close();
    },
    { scope: "worker", auto: false },
  ],

  sharedContextPage: async ({ sharedContext }, use) => {
    const page = await sharedContext.newPage();
    await use(page);
    await page.close();
  },
});
