# KubeDeck V1 微内核与 i18n 任务拆分

## 1. 目的

本文档把微内核与 i18n 的开发前置条件拆成可执行顺序。

## 2. 任务顺序

### Task A：内核合同定义

定义：
- 前端 page/menu/action/slot 合同形状
- 后端 capability/metadata/execution 合同形状

产出：
- 合同草案被接受

### Task B：内核归属映射

定义：
- shell 归属
- built-in capability 归属
- shared context 归属
- backend kernel 与 execution 归属

产出：
- 实现映射被接受

### Task C：i18n 最低运行时模型

定义：
- locale 偏好归属
- 文案访问边界
- 贡献元数据可本地化规则
- 新代码禁止继续扩散内联文案规则

产出：
- i18n 运行时模型被接受

### Task D：V1 扩展闸门

定义：
- 哪些功能实现必须暂停，直到微内核与 i18n 前置条件被满足
- 哪些工作可以并行继续

产出：
- V1 边界被接受

### Task E：首补丁规划

规划：
- 第一批面向内核的补丁范围
- 第一批 i18n 纪律补丁范围
- 防止意外扩范围的检查条件

产出：
- 首补丁范围与检查清单被接受

## 3. 推荐执行规则

在 Tasks A 到 D 被接受之前，不应继续扩大更深层功能实现。

Task E 应是进入代码前的最后一步文档准备。
