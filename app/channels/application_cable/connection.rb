module ApplicationCable
  class Connection < ActionCable::Connection::Base
    include Authentication::SessionLookup

    identified_by :current_user

    def connect
      self.current_user = find_verified_user
    end

    private
      def find_verified_user
        find_by_session_cookie || find_by_api_token || reject_unauthorized_connection
      end

      def find_by_session_cookie
        if verified_session = find_session_by_cookie
          verified_session.user
        end
      end

      def find_by_api_token
        if token = request.params[:token].presence
          User.active.find_by(api_token: token)
        end
      end
  end
end
