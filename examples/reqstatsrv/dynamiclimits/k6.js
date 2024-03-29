import{sleep} from "k6";
import http from "k6/http";


export default function() {
    let port = 7900
    /* Each 'client' will try to make around 1000 requests per second,
     * that are 60000 requests per minute. So, after each request
     * it will wait for 1/1000 seconds.
     */
    let reqPerSec = 1
    let params = {
        headers: {
            'X-Api-Key': '0000001'
        }
    }

    let endpoint = "http://192.168.1.20:" + port.toString() + "/endpoint_a/whatever"
    let response = http.get(endpoint, params)

    if(response.status === 200){
        console.log("Ok")
    } else if (response.status === 429 ){
        console.log("RATE LIMITED")
    } else {
        console.log("FAIL")
    }
    sleep(1/reqPerSec)
};
