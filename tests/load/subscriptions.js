import http from "k6/http";
import { check, sleep } from "k6";
import exec from "k6/execution";

const baseURL = __ENV.BASE_URL || "http://localhost:8080";
const userID = __ENV.USER_ID || "60601fee-2bf1-4721-ae6f-7636e79a0cba";
const rate = Number(__ENV.RATE || 0);

export const options = {
  scenarios:
    rate > 0
      ? {
          subscriptions_crud_total: {
            executor: "constant-arrival-rate",
            rate,
            timeUnit: "1s",
            duration: __ENV.DURATION || "1m",
            preAllocatedVUs: Number(__ENV.VUS || 50),
            maxVUs: Number(__ENV.MAX_VUS || 200),
          },
        }
      : {
          subscriptions_crud_total: {
            executor: "constant-vus",
            vus: Number(__ENV.VUS || 10),
            duration: __ENV.DURATION || "1m",
          },
        },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500", "p(99)<1000"],
  },
};

export function setup() {
  const res = http.get(`${baseURL}/health`, {
    tags: { name: "GET /health" },
  });
  const healthy = check(res, {
    "health status is 200": (r) => r.status === 200,
  });
  if (!healthy) {
    throw new Error(`service is not available at ${baseURL}; start it with make docker-up and check /health`);
  }
}

export default function () {
  const suffix = `${exec.vu.idInTest}-${exec.scenario.iterationInTest}-${Date.now()}`;
  const serviceName = `Load Test ${suffix}`;
  const headers = { "Content-Type": "application/json" };

  const createRes = http.post(
    `${baseURL}/api/v1/subscriptions`,
    JSON.stringify({
      service_name: serviceName,
      price: 400,
      user_id: userID,
      start_date: "07-2025",
      end_date: "12-2025",
    }),
    { headers, tags: { name: "POST /api/v1/subscriptions" } },
  );

  const subscriptionID = jsonValue(createRes, "id");
  const created = check(createRes, {
    "create status is 201": (r) => r.status === 201,
    "create returns id": () => Boolean(subscriptionID),
  });

  if (!created) {
    return;
  }

  const getRes = http.get(`${baseURL}/api/v1/subscriptions/${subscriptionID}`, {
    tags: { name: "GET /api/v1/subscriptions/{id}" },
  });
  check(getRes, {
    "get status is 200": (r) => r.status === 200,
    "get returns same id": (r) => jsonValue(r, "id") === subscriptionID,
  });

  const listRes = http.get(
    `${baseURL}/api/v1/subscriptions?user_id=${userID}&service_name=${encodeURIComponent(serviceName)}&limit=20&offset=0`,
    { tags: { name: "GET /api/v1/subscriptions" } },
  );
  check(listRes, {
    "list status is 200": (r) => r.status === 200,
    "list returns array": (r) => Array.isArray(jsonValue(r)),
  });

  const totalRes = http.get(
    `${baseURL}/api/v1/subscriptions/total?from=07-2025&to=12-2025&user_id=${userID}&service_name=${encodeURIComponent(serviceName)}`,
    { tags: { name: "GET /api/v1/subscriptions/total" } },
  );
  check(totalRes, {
    "total status is 200": (r) => r.status === 200,
    "total is correct": (r) => jsonValue(r, "total") === 2400,
  });

  const updateRes = http.put(
    `${baseURL}/api/v1/subscriptions/${subscriptionID}`,
    JSON.stringify({
      service_name: serviceName,
      price: 500,
      user_id: userID,
      start_date: "07-2025",
      end_date: "12-2025",
    }),
    { headers, tags: { name: "PUT /api/v1/subscriptions/{id}" } },
  );
  check(updateRes, {
    "update status is 200": (r) => r.status === 200,
    "update returns new price": (r) => jsonValue(r, "price") === 500,
  });

  const deleteRes = http.del(`${baseURL}/api/v1/subscriptions/${subscriptionID}`, null, {
    tags: { name: "DELETE /api/v1/subscriptions/{id}" },
  });
  check(deleteRes, {
    "delete status is 204": (r) => r.status === 204,
  });

  sleep(1);
}

function jsonValue(response, selector) {
  if (!response || response.status === 0 || !response.body) {
    return null;
  }

  try {
    return selector ? response.json(selector) : response.json();
  } catch {
    return null;
  }
}
