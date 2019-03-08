<?php

require "vendor/autoload.php";

$client = new Pb\chatServiceClient('localhost:40040', [
    'credentials' => Grpc\ChannelCredentials::createInsecure(),
]);


$stream = $client->subscribe(new \Google\Protobuf\GPBEmpty());
$messages = $stream->responses();
foreach ($messages as $msg) {
    var_dump($msg);
}
