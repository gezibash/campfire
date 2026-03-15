class Messages::BoostsController < ApplicationController
  include ActionView::RecordIdentifier

  before_action :set_message

  def index
  end

  def new
  end

  def create
    @boost = @message.boosts.create!(boost_params)

    broadcast_create
    redirect_to message_boosts_url(@message)
  end

  def destroy
    @boost = Current.user.boosts.find(params[:id])
    @boost.destroy!

    broadcast_remove
    redirect_to message_boosts_url(@message)
  end

  private
    def set_message
      @message = Current.user.reachable_messages.find(params[:message_id])
    end

    def boost_params
      params.require(:boost).permit(:content)
    end

    def broadcast_boosts
      @message.reload
      Turbo::StreamsChannel.broadcast_replace_to(
        [ @message.room, :messages ],
        target: dom_id(@message, :boosting),
        partial: "messages/boosts/boosts",
        locals: { message: @message },
        attributes: { maintain_scroll: true }
      )
    end

    alias_method :broadcast_create, :broadcast_boosts
    alias_method :broadcast_remove, :broadcast_boosts
end
