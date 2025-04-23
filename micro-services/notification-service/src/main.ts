import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import * as dotenv from 'dotenv';
import 'reflect-metadata';

// Load environment variables
dotenv.config();

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  
  // Add global prefix
  app.setGlobalPrefix('api/v1');
  
  // Start listening on the configured port
  const port = process.env.PORT || 3000;
  await app.listen(port);
  
  console.log(`Notification service running on http://localhost:${port}`);
}

bootstrap();
