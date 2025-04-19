import { Controller, Post, Body, Module, Injectable, OnModuleInit } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import {
  SNSClient,
  PublishCommand,
} from '@aws-sdk/client-sns';
import {
  SQSClient,
  ReceiveMessageCommand,
  DeleteMessageCommand,
} from '@aws-sdk/client-sqs';
import * as dotenv from 'dotenv';

dotenv.config();

@Injectable()
class NotificationService implements OnModuleInit {
  private snsClient: SNSClient;
  private sqsClient: SQSClient;
  private polling: boolean = true;

  constructor() {
    this.snsClient = new SNSClient({
      region: process.env.AWS_REGION,
      credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!,
      },
    });

    this.sqsClient = new SQSClient({
      region: process.env.AWS_REGION,
      credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!,
      },
    });
  }

  async onModuleInit() {
    this.pollMessages();
  }

  async sendEmail(to: string, subject: string, message: string) {
    const command = new PublishCommand({
      Message: `Subject: ${subject}\n\n${message}`,
      TopicArn: process.env.SNS_EMAIL_TOPIC_ARN,
      MessageAttributes: {
        recipient: {
          DataType: 'String',
          StringValue: to,
        },
      },
    });
    await this.snsClient.send(command);
  }

  async sendSMS(phoneNumber: string, message: string) {
    const command = new PublishCommand({
      Message: message,
      PhoneNumber: phoneNumber,
    });
    await this.snsClient.send(command);
  }

  async sendPush(deviceToken: string, message: string) {
    const command = new PublishCommand({
      Message: message,
      TopicArn: process.env.SNS_PUSH_TOPIC_ARN,
      MessageAttributes: {
        deviceToken: {
          DataType: 'String',
          StringValue: deviceToken,
        },
      },
    });
    await this.snsClient.send(command);
  }

  private async pollMessages() {
    const queueUrl = process.env.SQS_NOTIFICATION_QUEUE_URL!;
    console.log('SQS polling started...');
    while (this.polling) {
      const command = new ReceiveMessageCommand({
        QueueUrl: queueUrl,
        MaxNumberOfMessages: 5,
        WaitTimeSeconds: 10,
      });

      const response = await this.sqsClient.send(command);

      if (response.Messages) {
        for (const msg of response.Messages) {
          try {
            const body = JSON.parse(msg.Body!);
            const data = JSON.parse(body.Message || '{}');

            console.log('Received SQS Event:', data);

            if (data.type === 'email') {
              await this.sendEmail(data.to, data.subject, data.message);
            } else if (data.type === 'sms') {
              await this.sendSMS(data.phoneNumber, data.message);
            } else if (data.type === 'push') {
              await this.sendPush(data.deviceToken, data.message);
            }

            await this.sqsClient.send(
              new DeleteMessageCommand({
                QueueUrl: queueUrl,
                ReceiptHandle: msg.ReceiptHandle!,
              }),
            );
          } catch (err) {
            console.error('Error handling SQS message:', err);
          }
        }
      }
    }
  }
}

@Controller('notify')
class NotificationController {
  constructor(private readonly service: NotificationService) {}

  @Post('email')
  async email(@Body() body: { to: string; subject: string; message: string }) {
    await this.service.sendEmail(body.to, body.subject, body.message);
    return { status: 'email sent' };
  }

  @Post('sms')
  async sms(@Body() body: { phoneNumber: string; message: string }) {
    await this.service.sendSMS(body.phoneNumber, body.message);
    return { status: 'sms sent' };
  }

  @Post('push')
  async push(@Body() body: { deviceToken: string; message: string }) {
    await this.service.sendPush(body.deviceToken, body.message);
    return { status: 'push sent' };
  }
}

@Module({
  controllers: [NotificationController],
  providers: [NotificationService],
})
class AppModule {}

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  app.setGlobalPrefix(''); // so it's accessible from mallhive.com/
  await app.listen(3000);
  console.log(`Notification service ready at http://localhost:3000`);
}
bootstrap();
