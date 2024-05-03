# Logues

An Instant Messaging applicaton written in Golang.\
The main driver of the project is to learn, implement & showcase knowledge in the following fields.
* Programming Languages
    * Golang - Server & CLI Client
    * HTTP/CSS & JavaScript - Web Client

* Tools
    * Apache Kafka
    * RabbitMQ
    * Terraform
    * Azure

* Style of development
    * TDD
    * BDD

* Microservices Architecture

The project's (current) roadmap
1. Registration & Authentication of users. Access to a single global communication channel & messaging.
2. CI / CD Pipeline. Terraform. Azure.
3. Data consistency between user sessions on channels.
4. User operations, creation & management of Forums & channels.
5. Apache Kafka & RabbitMQ


Domain Terminology:
* User - registered & authenticated identity.
* Registeration - create new users & store credentials.
* Authentication - verifies user's credentials.
* Client - internel controller for user client based on WebSocket communication.
* Message - user generated data sent to other users.
* Channel - medium of communication where messages are sent.
* Forum - aggregation of users & channels with a ruleset.
* Ruleset - determines who & what actions can be done in the forum's context
* Event - An aggregate of operation & notification.
* Operations - Actions done by users which have an effect on other users, forums & channels.
* Notification - Information related to an operation other users performed and as relevence.
