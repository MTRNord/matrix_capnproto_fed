# What if Matrix was written using Cap'n'proto and a RPC federation API?

This question is what I am trying to proof by actually implementing this into synapse.

For simplicity this repo includes some go demo code for some of the details but
it also provides a suggested structure of folders for the proto files.

## Will there be an MSC?

Maybe but my hopes are either way very low that this will get mainline.

## Will there be a synapse PR?

Yes this is a major goal of this. It will also support fallback by design to the HTTP API
if the rpc API is not available.

## Will this work with $Loadbalancer?

It depends. Its doing RPC over unix or tcp sockets. So if your loadbalancer can handle that
it should work.

## Will this be faster?

Maybe. This is to be tested

## What will be signed?

The message that holds the signature map will be signed in its canonical binary form.
