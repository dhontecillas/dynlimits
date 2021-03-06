import http from "k6/http";

export default function() {
    let response = http.get("http://172.17.0.1:9876")
};
