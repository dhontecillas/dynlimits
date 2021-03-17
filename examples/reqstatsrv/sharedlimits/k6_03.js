import{sleep, check} from "k6";
import { Counter } from "k6/metrics"
import http from "k6/http";

let respOk = new Counter('response_ok')
let respRateLimited = new Counter('response_ratelimited')
let respOther = new Counter('response_other')

export default function() {
    let port = 7903
    /* Each 'client' will try to make around 1000 requests per second,
     * that are 60000 requests per minute. So, after each request
     * it will wait for 1/1000 seconds.
     */
    let reqPerSec = 1000
    let params = {
        headers: {
            'X-Api-Key': '000001'
        }
    }
    let endpoint = "http://192.168.1.20:" + port.toString() + "/foo"
    let response = http.get(endpoint, params)

    if(response.status === 200){
        respOk.add(1)
    } else if (response.status === 429 ){
        respRateLimited.add(1)
    } else {
        respOther.add(1)
    }
    sleep(1/reqPerSec)
};
