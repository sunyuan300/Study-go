# Commit Message作用

- 清晰地知道每个commit的变更内容。
- 可以基于规范化的Commit Message **生成Change Log**。
- 可以依据某些类型的Commit Message **触发构建或者发布流程**，比如当type类型为feat、fix才触发CI流程。
- **确定语义化版本的版本号**。比如fix类型可以映射为PATCH版本，feat类型可以映射为MINOR版本。带有BREAKING CHANGE的commit可以映射为MAJOR版本。