# accessdoor
This is the external service consumed by the Website/App . This service is responsible for calling all the internal services and formatting the data according to the consumer.
This service can be accessed by any consumer app/website/other services . 
This service has the business logic and validatation on the internal services and internal services will act like the datasources. 

# Endpoint 
GET /getuser - This endpoint publishes all the user information along with the historic event information. 

POST /updateuseraccess - This endpoint is used to update the user access.Only admin users can update the user access.

POST /authenticate - This endpoint is used to authenticate if the user has access to a door. If yes the event is saved in the events database.

# Internal Service Communication 
- users-go
- events-go


