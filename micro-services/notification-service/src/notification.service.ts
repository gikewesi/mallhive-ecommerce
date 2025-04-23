import { Injectable, OnModuleInit } from '@nestjs/common';
import {
  SNSClient,
  PublishCommand,
} from '@aws-sdk/client-sns';
import {
  SQSClient,
  ReceiveMessageCommand,
  DeleteMessageCommand,
} from '@aws-sdk/client-sqs';
import axios from 'axios';

@Injectable()
export class NotificationService implements OnModuleInit {
  private snsClient: SNSClient;
  private sqsClient: SQSClient;
  private polling = true;

  constructor() {
    this.snsClient = new SNSClient({
      region: process.env.AWS_REGION,
      credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
      },
    });

    this.sqsClient = new SQSClient({
      region: process.env.AWS_REGION,
      credentials: {
        accessKeyId: process.env.AWS_ACCESS_KEY_ID,
        secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
      },
    });
  }

  async onModuleInit(): Promise<void> {
    // Start polling for messages from SQS when the service initializes
    this.pollMessages();
    console.log('NotificationService initialized - SQS polling started');
  }

  async sendEmail(to: string, subject: string, message: string): Promise<void> {
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
    console.log(`Email sent to ${to}: ${subject}`);
  }

  async sendSMS(phoneNumber: string, message: string): Promise<void> {
    const command = new PublishCommand({
      Message: message,
      PhoneNumber: phoneNumber,
    });
    
    await this.snsClient.send(command);
    console.log(`SMS sent to ${phoneNumber}`);
  }

  async sendPush(deviceToken: string, message: string): Promise<void> {
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
    console.log(`Push notification sent to device ${deviceToken}`);
  }

  private async pollMessages(): Promise<void> {
    const queueUrl = process.env.SQS_NOTIFICATION_QUEUE_URL;

    while (this.polling) {
      try {
        const command = new ReceiveMessageCommand({
          QueueUrl: queueUrl,
          MaxNumberOfMessages: 5,
          WaitTimeSeconds: 10,
        });

        const response = await this.sqsClient.send(command);

        if (response.Messages) {
          for (const msg of response.Messages) {
            try {
              const body = JSON.parse(msg.Body);
              const data = JSON.parse(body.Message || '{}');

              console.log('Received SQS Event:', data);

              // Handle based on type
              switch (data.type) {
                case 'email':
                  await this.sendEmail(data.to, data.subject, data.message);
                  break;

                case 'sms':
                  await this.sendSMS(data.phoneNumber, data.message);
                  break;

                case 'push':
                  await this.sendPush(data.deviceToken, data.message);
                  break;

                case 'payment':
                  await this.handlePaymentEvent(data);
                  break;

                case 'order':
                  await this.handleOrderEvent(data);
                  break;

                case 'user':
                  await this.handleUserEvent(data);
                  break;

                default:
                  console.warn('Unknown event type:', data.type);
              }

              // Delete the message after processing
              await this.sqsClient.send(
                new DeleteMessageCommand({
                  QueueUrl: queueUrl,
                  ReceiptHandle: msg.ReceiptHandle,
                }),
              );
            } catch (err) {
              console.error('Error handling SQS message:', err);
            }
          }
        }
      } catch (err) {
        console.error('Error polling SQS:', err);
        // Wait before retrying to avoid tight loop in case of persistent errors
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }
  }

  private async handlePaymentEvent(data: any): Promise<void> {
    const url = `${process.env.PAYMENT_SERVICE_URL}/status`;

    try {
      // Notify payment service about the status update
      await axios.post(url, {
        paymentId: data.paymentId,
        status: data.status,
      });

      // Determine message based on payment status
      let message: string;
      if (data.status === 'completed') {
        message = `Payment completed successfully for transaction ID: ${data.paymentId}`;
      } else if (data.status === 'failed') {
        message = `Payment failed for transaction ID: ${data.paymentId}. Please try again.`;
      } else if (data.status === 'canceled') {
        message = `Payment canceled for transaction ID: ${data.paymentId}.`;
      } else {
        message = `Payment ${data.status} for transaction ID: ${data.paymentId}`;
      }

      // Send email notification
      await this.sendEmail(data.email, 'Payment Notification', message);
      
      // If phone number is provided, also send SMS
      if (data.phoneNumber) {
        await this.sendSMS(data.phoneNumber, message);
      }
    } catch (err) {
      console.error('Error handling payment event:', err);
    }
  }

  private async handleOrderEvent(data: any): Promise<void> {
    const url = `${process.env.ORDER_SERVICE_URL}/notify`;

    try {
      // Notify order service
      await axios.post(url, {
        orderId: data.orderId,
        userId: data.userId,
        status: data.status || 'placed',
      });

      // Send notification to user about order
      const message = `Your order #${data.orderId} has been ${data.status || 'placed'} successfully.`;
      await this.sendEmail(data.email, 'Order Notification', message);
      
      // If phone number is provided, also send SMS
      if (data.phoneNumber) {
        await this.sendSMS(data.phoneNumber, message);
      }
    } catch (err) {
      console.error('Error handling order event:', err);
    }
  }

  private async handleUserEvent(data: any): Promise<void> {
    const url = `${process.env.USER_SERVICE_URL}/activity`;

    try {
      // Notify user service
      await axios.post(url, {
        userId: data.userId,
        action: data.action,
      });

      // Send notification to user
      let message: string;
      let subject: string;
      
      if (data.action === 'login') {
        subject = 'Login Notification';
        message = 'You have successfully logged in to your account.';
      } else if (data.action === 'register') {
        subject = 'Welcome to Our Platform';
        message = 'Your account has been successfully created.';
      } else {
        subject = 'Account Activity';
        message = `Your account has been updated: ${data.action}`;
      }

      await this.sendEmail(data.email, subject, message);
      
      // If phone number is provided, also send SMS
      if (data.phoneNumber) {
        await this.sendSMS(data.phoneNumber, message);
      }
    } catch (err) {
      console.error('Error handling user event:', err);
    }
  }
}
