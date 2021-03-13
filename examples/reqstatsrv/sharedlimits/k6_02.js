import{sleep, check} from "k6";
import http from "k6/http";

export default function() {
    let port = 7902
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
    check(response, {
        'non blocked': (res) => res.status === 200
         })
    sleep(1/reqPerSec)
};
