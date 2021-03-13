import{sleep, check} from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics"

let respOk = new Counter('response_ok')
let respRateLimited = new Counter('response_ratelimited')
let respOther = new Counter('response_other')

export default function() {
    // the raw dummy server does not take into account the
    // api keys nor the path. We just setup this script
    // as close as possible to the other one.
    let apikey = ('000000' + __VU).slice(-6)
    let params = {
        headers: {
            'X-Api-Key': apikey
        }
    }
    let endpoint = "http://192.168.1.20:9876/endpoint_1"
    let response = http.get(endpoint, params)

    if(response.status === 200){
        respOk.add(1)
    } else if (response.status === 429 ){
        respRateLimited.add(1)
    } else {
        respOther.add(1)
    }
};
