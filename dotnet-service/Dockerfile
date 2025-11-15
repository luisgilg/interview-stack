# syntax=docker/dockerfile:1
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

COPY src/DotnetService.csproj ./
RUN dotnet restore DotnetService.csproj
COPY src/ ./
RUN dotnet publish DotnetService.csproj -c Release -o /app/publish /p:UseAppHost=false

FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app

COPY --from=build /app/publish ./
COPY wait-for.sh ./wait-for.sh
RUN apt-get update && apt-get install -y --no-install-recommends netcat-traditional \
    && rm -rf /var/lib/apt/lists/*

ENV DOTNET_SERVICE_PORT=8083
EXPOSE 8083

ENTRYPOINT ["/bin/sh", "-c", "./wait-for.sh ${POSTGRES_HOST:-postgres} ${POSTGRES_PORT:-5432} ./wait-for.sh ${MONGO_HOST:-mongo} ${MONGO_PORT:-27017} ./wait-for.sh ${REDIS_HOST:-redis} ${REDIS_PORT:-6379} dotnet DotnetService.dll"]
