require "zlib"

module Users::AvatarsHelper
  AVATAR_COLORS = %w[
    #AF2E1B #CC6324 #3B4B59 #BFA07A #ED8008 #ED3F1C #BF1B1B #736B1E #D07B53
    #736356 #AD1D1D #BF7C2A #C09C6F #698F9C #7C956B #5D618F #3B3633 #67695E
  ]

  def avatar_background_color(user)
    AVATAR_COLORS[Zlib.crc32(user.to_param) % AVATAR_COLORS.size]
  end

  def avatar_tag(user, **options)
    link_to user_path(user), title: user.title, class: "btn avatar", data: { turbo_frame: "_top" } do
      image_tag fresh_user_avatar_path(user), aria: { hidden: "true" }, size: 48, **options
    end
  end

  def presence_dot(user)
    return "".html_safe unless user.respond_to?(:last_seen_at) && user.last_seen_at.present?

    age = Time.current - user.last_seen_at
    if age < 5.minutes
      tag.span class: "avatar__presence avatar__presence--online", title: "Online"
    elsif age < 30.minutes
      tag.span class: "avatar__presence avatar__presence--away", title: "Away"
    else
      "".html_safe
    end
  end

  def presence_color(user)
    return nil unless user.respond_to?(:last_seen_at) && user.last_seen_at.present?

    age = Time.current - user.last_seen_at
    if age < 5.minutes
      "#22c55e"
    elsif age < 30.minutes
      "#eab308"
    end
  end

  def last_seen_text(user)
    return nil unless user.respond_to?(:last_seen_at) && user.last_seen_at.present?

    age = Time.current - user.last_seen_at
    if age < 1.minute
      "Online now"
    elsif age < 5.minutes
      "Active #{time_ago_in_words(user.last_seen_at)} ago"
    elsif age < 30.minutes
      "Away \u00b7 #{time_ago_in_words(user.last_seen_at)} ago"
    else
      "Last seen #{time_ago_in_words(user.last_seen_at)} ago"
    end
  end
end
