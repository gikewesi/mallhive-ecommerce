import { Controller, Post, Body, Module, Injectable } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { SNSClient, PublishCommand } from '@aws-sdk/client-sns';
import * as dotenv from 'dotenv';

dotenv.config();

@Injectable()
class NotificationService {
  private snsClient: SNSClient;

  constructor() {
    this.snsClient = new SNSClient({
      region: process.env.AWS_REGION || 'us-east-1',
      credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID || '',
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY || '',
      },
    });
  }

  async sendEmail(to: string, subject: string, message: string) {
    // Using SNS to simulate email
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
    // Simulated push notification via SNS
    const command = new PublishCommand({
      Message: `Push to ${deviceToken}: ${message}`,
      TopicArn: process.env.SNS_PUSH_TOPIC_ARN,
    });
    await this.snsClient.send(command);
  }
}

@Controller('/')
class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Post('notify/email')
  async sendEmail(@Body() body: { to: string; subject: string; message: string }) {
    await this.notificationService.sendEmail(body.to, body.subject, body.message);
    return { status: 'Email sent' };
  }

  @Post('notify/sms')
  async sendSMS(@Body() body: { phoneNumber: string; message: string }) {
    await this.notificationService.sendSMS(body.phoneNumber, body.message);
    return { status: 'SMS sent' };
  }

  @Post('notify/push')
  async sendPush(@Body() body: { deviceToken: string; message: string }) {
    await this.notificationService.sendPush(body.deviceToken, body.message);
    return { status: 'Push notification sent' };
  }
}

@Module({
  controllers: [NotificationController],
  providers: [NotificationService],
})
class AppModule {}

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  app.setGlobalPrefix(''); // so endpoints are at https://mallhive.com/notify/*
  await app.listen(3000);
  console.log(`Notification service is live at http://localhost:3000`);
}
bootstrap();
