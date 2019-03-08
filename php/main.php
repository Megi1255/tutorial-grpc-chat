<?php

require "vendor/autoload.php";

$client = new Pb\chatServiceClient('localhost:40040', [
    'credentials' => Grpc\ChannelCredentials::createInsecure(),
]);


$stream = $client->subscribe(new \Google\Protobuf\GPBEmpty());
$messages = $call->respones();
foreach ($msg as $messages) {
    var_dump($msg);
}