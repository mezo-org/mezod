# RFC-4: Native non-Bitcoin asset Bridge

## Background
[RFC-3: Bridging non-Bitcoin assets to Mezo](./rfc-3.md) was the first attempt 
to describe the mechanism to support the operation of bridging non-Bitcoin
assets to Mezo minimizing the fragmentation of liquidity . While all of the
concepts presented in RFC-3 are still technically correct, the RFC-3 proposal is 
based on the existence of a Bridging Partner ready to perform to-Mezo bridging 
on day one, and this assumption is quite risky in practice.

RFC-4 proposes an extension to the Bitcoin bridge mechanism described in
[RFC-2: Bridging Bitcoin to Mezo](./rfc-2.md) to support a small, selected group
of non-Bitcoin assets in the native bridge. The non-Bitcoin assets not supported
by the native bridge described in RFC-4 will be bridgable in the future, most
probably using a mechanism described in RFC-3.

