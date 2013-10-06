require 'sinatra'

get '/' do
  backends = {
    8080 => 'production',
    8081 => 'testing',
  }

  sleep 3 if backends[request.port] == 'testing'

  logger.info("sandbox: #{request.env['HTTP_X_DELTA_SANDBOX'] ? 1 : 0}")
  "#{backends[request.port]}"
end
