
import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const BASE_URL = (__ENV.BASE_URL || "http://localhost:8080")
    .replace(/\/$/, "");

const roomsListDuration = new Trend("rooms_list_duration", true);
const slotsListDuration = new Trend("slots_list_duration", true);
const errorRate = new Rate("api_error_rate");

export const options = {
    summaryTrendStats: [
        "avg",
        "min",
        "med",
        "max",
        "p(90)",
        "p(95)",
        "p(99)",
    ],

    stages: [
        { duration: "10s", target: 2 },
        { duration: "30s", target: 5 },
        { duration: "10s", target: 0 },
    ],

    thresholds: {
        http_req_failed: ["rate<0.01"],
        checks: ["rate>0.99"],
        api_error_rate: ["rate<0.01"],

        rooms_list_duration: [
            "p(95)<500",
            "p(99)<1000",
        ],

        slots_list_duration: [
            "p(95)<500",
            "p(99)<1000",
        ],
    },
};

export function setup() {
    const loginResponse = http.post(
        `${BASE_URL}/dummyLogin`,
        JSON.stringify({ role: "user" }),
        {
            headers: {
                "Content-Type": "application/json",
            },
        },
    );

    if (loginResponse.status !== 200) {
        throw new Error(
            `dummyLogin failed: HTTP ${loginResponse.status}`,
        );
    }

    const token = loginResponse.json("token");

    if (!token) {
        throw new Error("dummyLogin returned an empty token");
    }

    const headers = authHeaders(token);

    let roomID = __ENV.ROOM_ID || "";

    if (!roomID) {
        const roomsResponse = http.get(
            `${BASE_URL}/rooms/list`,
            { headers },
        );

        if (roomsResponse.status !== 200) {
            throw new Error(
                `Cannot get rooms: HTTP ${roomsResponse.status}`,
            );
        }

        const rooms = roomsResponse.json("rooms");

        if (!Array.isArray(rooms) || rooms.length === 0) {
            throw new Error(
                "No rooms found. Create a room or pass ROOM_ID.",
            );
        }

        roomID = rooms[0].id;
    }

    let testDate = __ENV.TEST_DATE || "";

    if (!testDate) {
        const tomorrow = new Date();
        tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
        testDate = tomorrow.toISOString().slice(0, 10);
    }

    if (!/^\d{4}-\d{2}-\d{2}$/.test(testDate)) {
        throw new Error("TEST_DATE must use YYYY-MM-DD format");
    }

    console.log(`BASE_URL: ${BASE_URL}`);
    console.log(`ROOM_ID: ${roomID}`);
    console.log(`TEST_DATE: ${testDate}`);

    return {
        token,
        roomID,
        testDate,
    };
}

export default function (data) {
    const headers = authHeaders(data.token);

    const roomsResponse = http.get(
        `${BASE_URL}/rooms/list`,
        { headers },
    );

    const roomsOK = check(roomsResponse, {
        "rooms: HTTP 200": (r) => r.status === 200,
        "rooms: valid array": (r) => {
            try {
                return Array.isArray(r.json("rooms"));
            } catch {
                return false;
            }
        },
    });

    roomsListDuration.add(roomsResponse.timings.duration);
    errorRate.add(!roomsOK);

    const slotsResponse = http.get(
        `${BASE_URL}/rooms/${data.roomID}/slots/list?date=${data.testDate}`,
        { headers },
    );

    const slotsOK = check(slotsResponse, {
        "slots: HTTP 200": (r) => r.status === 200,
        "slots: valid array": (r) => {
            try {
                return Array.isArray(r.json("slots"));
            } catch {
                return false;
            }
        },
    });

    slotsListDuration.add(slotsResponse.timings.duration);
    errorRate.add(!slotsOK);

    sleep(0.1);
}

function authHeaders(token) {
    return {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
    };
}
