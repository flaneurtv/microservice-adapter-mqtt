How to divide things up into mqtt-service-adapter repository and the actual service repositories:

The ./service-adapter dir represents the final mqtt-service-adapter submodule.
Everything else shall be part of the actual service repository e.g. the ticker-service or the tick_responder service.
Both these service repositories will contain therefore contain the mqtt-service-adapter repository as a git submodule under ./service-adapter dir.
And of course each of them will contain their respective processor file and a Dockerfile adjusted to support the programming language used in the processor file.
