atlas改造

1. FieldSelection 支持_nf/_nnf/!xxx/yyy._nnf/yyy.!xxx字段，寓意为 嵌套实体字段 以及 非嵌套实体字段 以及 排除xxx字段 以及 yyy下的非嵌套实体字段 以及 yyy下排除xxx字段：
- 在gorm fields.go里的handlePreloads里修改支持这俩特性
- 在gateway fields.go里的doRetainFields里修改支持这俩特性

2. FieldSelection 支持在field级别或者method级别设置默认值，使得能很方便的支持非自动preload等行为：
- 生成gorm代码时候对应request pb model生成DefaultValForQueryKeys方法 return map[string]string
- 在gateway的ClientUnaryInterceptor方法里CollectionOps相关逻辑里去检测如果某Query Key为空，则使用默认值。
- 在gateway fields.go里的doRetainFields里由于也依赖于req.URL.Query()，然后处理FieldSelection，所以也需要处理默认值逻辑，如果不好拿取，则尝试context

3. Field 支持在Field级别设置Invisible：
- 在对应的pb定义中，生成InvisibleFields方法，returns字段名列表即可，最好是驼峰和蛇形两份
- 对于Read和List请求体生成时候检测对应实体或者其N层嵌套实体是否存在InvisibleFields，如果存在，则整理进一个NecessaryFields方法，且要求请求提必须存在_fileds字段，返回例子 !f1,!f2,n1.!f3,n2.f5
- 类似第二点，在gateway的ClientUnaryInterceptor方法里将NecessaryFields的结果追加进_fields参数里，且doRetainFields也要对应处理

4. FieldSelection 支持[col~order~offset~limit]特性，寓意为以col的order序的offset的limit行：
- Field里加个标识这玩意的参数，在Parse方法里处理下
- 在has_many的Preload的回调里根据这些数据来执行相应行为

5. Field 支持在Field级别设置不可编辑：
- 对应的pb定义中，生成DenyUpdateFields方法，returns字段名列表即可，最好是驼峰和蛇形两份
- Update 请求在头部拿到pb model所有的Fields名称作为UpdateMasks，然后根据fields参数以及DenyUpdateFields结果对该去除的去除，最终只执行DefaultPatchXXX方法。

6. 为了更易用：
- DefaultStrictUpdateXXX的clean has_many的机制去除
- Create和Update在执行gorm Create以及Save方法之前要忽略嵌套实体，因为其用到的场景极少，而且容易出问题
- 根据上一条，DefaultPatchXXX 头部 Read 原有数据这个行为，默认给予 FieldSelection 为_nnf，能直接忽略嵌套实体的返回。
- UpdateMasks 如果有嵌套定义，直接报错
- Create和Update的请求如果response里有返回实体，则request里必须提供 FieldSelection 参数，且最终会重新执行一次read并返回


首要改造顺序：
1 5 6