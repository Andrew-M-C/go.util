---
name: date-query
description: >-
  Get the current local date and time using the shell `date` command.
  Use when the user asks about the current time or date.
---
# date-query: 获取当前日期时间

通过 shell 命令获取当前日期和时间。

## Usage

直接运行以下命令获取完整日期时间：

```bash
date
```

如需 24 小时制 HH:MM:SS 格式：

```bash
date '+%H:%M:%S'
```

运行命令后，将结果整理后用自然语言告知用户。
