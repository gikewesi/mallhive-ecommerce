import { Controller, Post, Body } from '@nestjs/common';
import { NotificationService } from './notification.service';

@Controller('notify')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Post('email')
  async sendEmail(
    @Body() body: { to: string; subject: string; message: string },
  ): Promise<{ status: string }> {
    await this.notificationService.sendEmail(body.to, body.subject, body.message);
    return { status: 'email sent' };
  }

  @Post('sms')
  async sendSMS(
    @Body() body: { phoneNumber: string; message: string },
  ): Promise<{ status: string }> {
    await this.notificationService.sendSMS(body.phoneNumber, body.message);
    return { status: 'sms sent' };
  }

  @Post('push')
  async sendPush(
    @Body() body: { deviceToken: string; message: string },
  ): Promise<{ status: string }> {
    await this.notificationService.sendPush(body.deviceToken, body.message);
    return { status: 'push notification sent' };
  }
}
