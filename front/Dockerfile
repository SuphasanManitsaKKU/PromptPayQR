# Base image
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy only necessary files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install --production

# Copy all source code except node_modules
COPY . .

# Expose port
EXPOSE 3000

# Command to start the Next.js application
CMD ["npm", "run", "dev"]
