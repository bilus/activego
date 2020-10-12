module ApplicationCable
  class Connection < ActionCable::Connection::Base
    identified_by :user

    def connect
      self.user = request.params['user']
      reject_unauthorized_connection unless self.user
    end
  end
end
