<?php
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chat-gateway.proto

namespace GPBMetadata;

class ChatGateway
{
    public static $is_initialized = false;

    public static function initOnce() {
        $pool = \Google\Protobuf\Internal\DescriptorPool::getGeneratedPool();

        if (static::$is_initialized == true) {
          return;
        }
        \GPBMetadata\Google\Api\Annotations::initOnce();
        \GPBMetadata\Google\Protobuf\GPBEmpty::initOnce();
        $pool->internalAddGeneratedFile(hex2bin(
            "0a98020a12636861742d676174657761792e70726f746f120270621a1b67" .
            "6f6f676c652f70726f746f6275662f656d7074792e70726f746f22230a07" .
            "4d657373616765120a0a026964180120012809120c0a0474657874180220" .
            "01280932b3010a0b6368617453657276696365124b0a0473656e64120b2e" .
            "70622e4d6573736167651a162e676f6f676c652e70726f746f6275662e45" .
            "6d707479221e82d3e493021822132f76312f636861747365727665722f73" .
            "656e643a012a12570a0973756273637269626512162e676f6f676c652e70" .
            "726f746f6275662e456d7074791a0b2e70622e4d657373616765222382d3" .
            "e493021d22182f76312f636861747365727665722f737562736372696265" .
            "3a012a3001620670726f746f33"
        ), true);

        static::$is_initialized = true;
    }
}

