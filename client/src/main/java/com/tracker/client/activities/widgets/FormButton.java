package com.tracker.client.activities.widgets;

import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;

import elemental.events.Event;
import elemental.events.EventListener;
import elemental.html.ButtonElement;

public class FormButton extends Component {
  public interface ClickHandler {
    void handleClick();
  }

  private ButtonElement button;
  private String title = "";
  private String name = "";
  private ClickHandler clickHandler;

  @Override
  protected void createDom() {
    button = DomUtils.createDom(doc.createButtonElement(), "btn btn-primary");
    button.setInnerText(title);
    button.setName(name);
    
    decorateInternal(button);
  }

  public FormButton setTitle(String title) {
    this.title = title;
    return this;
  }

  public FormButton setName(String name) {
    this.name = name;
    return this;
  }

  @Override
  public void enterDocument() {
    super.enterDocument();

    addHandlerRegistration(button.addEventListener(Event.CLICK, new EventListener() {
      @Override
      public void handleEvent(Event evt) {
        evt.stopPropagation();
        evt.preventDefault();
        if (clickHandler != null) {
          clickHandler.handleClick();
        }
      }
    }, false));
  }

  public void onClick(ClickHandler handler) {
    clickHandler = handler;
  }
}
