import{sleep, check} from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics"

let respOk = new Counter('response_ok')
let respRateLimited = new Counter('response_ratelimited')
let respOther = new Counter('response_other')

export default function() {
    let apikey = ('000000' + __VU).slice(-6)
    let params = {
        headers: {
            'X-Api-Key': apikey
        }
    }
    let endpoint = "http://192.168.1.20:7900/endpoint_" + (
        __ITER % (100- (__VU % 80)))
    let response = http.get(endpoint, params)

    if(response.status === 200){
        respOk.add(1)
    } else if (response.status === 429 ){
        respRateLimited.add(1)
    } else {
        respOther.add(1)
        console.log("res: " + response.status + JSON.stringify(response))
    }
};


