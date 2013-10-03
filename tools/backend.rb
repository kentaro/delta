require 'sinatra'

get '/' do
  backends = {
    8080 => 'production',
    8081 => 'testing',
  }

  "#{backends[request.port]}"
end

