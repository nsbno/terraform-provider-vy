resource "aws_sns_topic" "trigger" {
  name = "my-cool-topic.fifo"

  fifo_topic = true
}

resource "aws_sns_topic" "pipeline" {
  name = "my-other-cool-topic.fifo"

  fifo_topic = true
}

resource "vy_deployment_account" "this" {
  topics = {
    trigger_events  = aws_sns_topic.trigger.arn
    pipeline_events = aws_sns_topic.pipeline.arn
  }
}
