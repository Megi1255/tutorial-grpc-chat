<?php
// GENERATED CODE -- DO NOT EDIT!

namespace Pb;

/**
 */
class chatServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Pb\Message $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     */
    public function send(\Pb\Message $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/pb.chatService/send',
        $argument,
        ['\Google\Protobuf\GPBEmpty', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Google\Protobuf\GPBEmpty $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     */
    public function subscribe(\Google\Protobuf\GPBEmpty $argument,
      $metadata = [], $options = []) {
        return $this->_serverStreamRequest('/pb.chatService/subscribe',
        $argument,
        ['\Pb\Message', 'decode'],
        $metadata, $options);
    }

}
