module Api
  module V1
    class AvatarsController < BaseController
      include ActiveStorage::Streaming

      def show
        user = User.active.find_by(id: params[:user_id])
        unless user
          head :not_found
          return
        end

        expires_in 30.minutes, public: true, stale_while_revalidate: 1.week

        if user.avatar.attached?
          variant = user.avatar.variant(resize_to_limit: [ 512, 512 ], format: :webp).processed
          send_file ActiveStorage::Blob.service.path_for(variant.key), content_type: "image/webp", disposition: :inline
        else
          render_initials(user)
        end
      end

      private
        def render_initials(user)
          colors = %w[#E06B56 #5AB0E5 #6BBE6B #E0A156 #B56BE0 #56B0B0 #E05690 #90B056]
          color = colors[Zlib.crc32(user.id.to_s) % colors.length]
          initials = user.name.to_s.split.map { |w| w[0] }.first(2).join.upcase
          initials = "?" if initials.blank?

          svg = <<~SVG
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="100" height="100">
              <rect width="100" height="100" rx="10" fill="#{color}"/>
              <text x="50" y="55" text-anchor="middle" dominant-baseline="central"
                    font-family="system-ui, sans-serif" font-size="40" font-weight="600" fill="white">#{initials}</text>
            </svg>
          SVG

          send_data svg, type: "image/svg+xml", disposition: :inline
        end
    end
  end
end
