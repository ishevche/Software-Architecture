# Homework 3: Microservices with Hazelcast

Author: **Shevchenko Ivan**

GitHub: [https://github.com/ishevche/Software-Architecture/tree/micro_hazelcast/HW3](https://github.com/ishevche/Software-Architecture/tree/micro_hazelcast/HW3)

## Description

In this homework, I have added a Hazelcast cluster to a logging microservice
to have an ability to scale it and have consistent data, independent of the
instance of the service that the request was sent to.

## Usage

To run the application do the following steps:

1. Clone the repository
2. Move to the project root directory
3. Move to the `HW3` directory
4. Run `docker compose up` to run all services alongside all Hazelcast
   nodes and Hazelcast Manager Center
5. Facade service is now listening the port `8080`
6. Request could be made on the endpoint `http://localhost:8080/facade_service`
7. Hazelcast Managing server is running on `http://localhost:8090`

## Results

First of all, I have started the cluster using:

![Starting the service][compose_up]

We can assure that there is no data in freshly started cluster sending a `GET`
request:

![Response form the application for GET request immediately after start][empty_get_resp]

And we can see logs produced by this request:

![Logs for GET request immediately after start][empty_get_logs]

Then I have sent 10 `POST` requests with messages. Here are the logs produced
by these requests:

![Logs for POST requests][post_logs]

On the screenshot, we can see that requests were sent to different logging 
services, and each one printed the received message in the logs.

After that I have made a `GET` request, and received the following response:

![Response for GET request after POST requests are done][get_resp]

And this request produced last three lines of the following logs:

![Logs for GET requests after POST requests are done (last three lines)][get_logs]

After I have stopped two of three logging services alongside the respective
nodes of the Hazelcast:

![Stopping two of three logging services alongside corresponding Hazelcast nodes][stopping]

After stopping services I have repeated the `GET` request, and it resulted in 
the same output:

![Response for GET request after stopping logging services][stp:get_resp]

The only thing is different here is slightly higher response time. The reason
for this becomes transparent when we look at the logs produced by this request:

![Logs for GET request after stopping logging services][stp:get_logs]

Here we can see that `facade_service` tries to access `logging_service` but
connection failed, as it tried to access the service that was shut down. I have
set the number of retries to `5`, so it can take even more retries to get the
data, but nevertheless still get it:

![Response for GET request with more retries after stopping services][stp:get_more_resp]

We can see that response time is even higher than for the previous one, and
the reason is that it took even more retries, as we can see from logs:

![Logs for GET request with more retries after stopping logging services][stp:get_more_logs]

There also can be a situation when none of the requests succeeds, then the 
request fails and returns the error message:

![Response for failed GET request after stopping services][stp:get_fail_resp]

And the following is printed to the logs:

![Logs for failed GET request after stopping logging services][stp:get_fail_logs]

[compose_up]: img/cluster_started.png "Starting the service"
[empty_get_resp]: img/empty_get.png "Response form the application for GET request immediately after start"
[empty_get_logs]: img/empty_get_logs.png "Logs for GET request immediately after start"
[post_logs]: img/post_logs.png "Logs for POST requests"
[get_resp]: img/get_response.png "Response for GET request after POST requests are done"
[get_logs]: img/get_logs.png "Logs for GET requests after POST requests are done"
[stopping]: img/stopping_two_loggers.png "Stopping two of three logging services alongside corresponding Hazelcast nodes"
[stp:get_resp]: img/stopped_get_response.png "Response for GET request after stopping logging services"
[stp:get_logs]: img/stopped_get_logs.png "Logs for GET request after stopping logging services"
[stp:get_more_resp]: img/stopped_get_more_retries_response.png "Response for GET request with more retries after stopping services"
[stp:get_more_logs]: img/stopped_get_more_retries_logs.png "Logs for GET request with more retries after stopping logging services"
[stp:get_fail_resp]: img/stopped_get_failed_response.png "Response for failed GET request after stopping services"
[stp:get_fail_logs]: img/stopped_get_failed_logs.png "Logs for failed GET request after stopping logging services"
