# BloxID

This package, bloxid, implements typed guids for identifying resource objects globally in the system. Resources are not specific to services/applications but must contain an entity type.

Typed guids have the advantage of being globally unique, easily readable, and strongly typed for authorization and logging. The trailing characters provide sufficient entropy to make each resource universally unique.

bloxid package provides methods for generating and parsing versioned typed guids.

bloxid supports different schemes for the unique id portion of the bloxid
- `extrinsic`
- `hashid`
- `random`


## Scheme - extrinsic

create bloxid from extrinsic id
```golang
v0, err := NewV0("",
            WithEntityDomain("iam"),
            WithEntityType("user"),
            WithRealm("us-com-1"),
            WithExtrinsic("123456"),
        )
// v0.String(): blox0.iam.user.us-com-1.ivmfiurrgiztinjweaqcaiba
// v0.Decoded(): "123456"
```

parse bloxid to retrieve extrinsic id
```golang
parsed, err := NewV0("blox0.iam.user.us-com-1.ivmfiurrgiztinjweaqcaiba")
// v0.Decoded(): "123456"
```


## Scheme - hashid

create bloxid from hashid
```golang
v0, err := NewV0("",
            WithEntityDomain("infra"),
            WithEntityType("host"),
            WithRealm("us-com-1"),
            WithHashIDInt64(1),
            WithHashIDSalt("test"),
        )
// v0.String(): blox0.infra.host.us-com-1.jbeuiwrsmq3tkmzwmuzwcojsmrqwemrtgy3tqzbvhbsdizjvhe2dkn3cgzrdizlb
// v0.HashIDInt64(): 1
```

parse bloxid to retrieve hashid
```golang
parsed, err := NewV0("blox0.infra.host.us-com-1.jbeuiwrsmq3tkmzwmuzwcojsmrqwemrtgy3tqzbvhbsdizjvhe2dkn3cgzrdizlb",
    WithHashIDSalt("test")
)
// v0.HashIDInt64(): 1
```


## Scheme - random

create bloxid from generated random id
```golang
v0, err := NewV0("",
            WithEntityDomain("iam"),
            WithEntityType("group"),
            WithRealm("us-com-1"),
        )
// v0.String(): blox0.iam.group.us-com-1.tshwyq3mfkgqqcfa76a5hbr2uaayzw3h
// v0.Decoded(): "9c8f6c436c2a8d0808a0ff81d3863aa0018cdb67"
```

parse bloxid to retrieve extrinsic id
```golang
parsed, err := NewV0("blox0.iam.group.us-com-1.tshwyq3mfkgqqcfa76a5hbr2uaayzw3h")
// v0.Decoded(): "9c8f6c436c2a8d0808a0ff81d3863aa0018cdb67"
```

