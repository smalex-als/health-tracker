package com.tracker.client.activities.widgets;

import java.util.logging.Logger;

import com.tracker.client.controls.Component;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.StringUtils;
import com.tracker.client.util.StyleUtils;
import com.google.gwt.user.client.DOM;

import elemental.dom.Element;
import elemental.events.Event;
import elemental.events.EventListener;
import elemental.html.InputElement;
import elemental.html.LabelElement;

public class FormInputText extends Component {
  private static final Logger log = Logger.getLogger(FormInputText.class.getName());
  private Element rootEl;
  private InputElement input;
  private LabelElement label;
  private Element feedbackEl;
  private String title = "";
  private String name = "";
  private String type = "";
  private String autocomplete = "off";
  private String placeholder = "";
  private boolean required;
  private boolean spellcheck;
  private boolean grid;

  public FormInputText() {
  }

  @Override
  protected void createDom() {
    label = DomUtils.createDom(doc.createLabelElement());
    label.setTextContent(title);
    if (grid) {
      label.setClassName("col-form-label col-2");
    } else {
      label.setClassName("col-form-label");
    }

    input = DomUtils.createDom(doc.createInputElement(), "form-control");
    input.setId(DOM.createUniqueId());

    input.setType(type);
    input.setName(name);
    input.setMaxLength(255);
    if (StringUtils.hasLength(placeholder)) {
      input.setPlaceholder(placeholder);
    }
    if (StringUtils.hasLength(autocomplete)) {
      input.setAutocomplete(autocomplete);
    }
    if (!spellcheck) {
      input.setSpellcheck(spellcheck);
    }

    label.setHtmlFor(input.getId());

    if (grid) {
      rootEl = DomUtils.createDom(doc.createDivElement(), "form-group row", 
          label,
          DomUtils.createDom(doc.createDivElement(), "col-10", input));

    } else {
      rootEl = DomUtils.createDom(doc.createDivElement(), "form-group", label, input);
    }

    decorateInternal(rootEl);
    feedbackEl = DomUtils.createDom(doc.createDivElement(), "form-control-feedback");
  }

  @Override
  protected void decorateInternal(Element element) {
    super.decorateInternal(element);
  }

  public String getValue() {
    return StringUtils.trimToEmpty(input.getValue());
  }

  public void setValue(String value) {
    input.setValue(StringUtils.trimToEmpty(value));
  }

  public double getValueDouble() {
    return safeParseDouble(StringUtils.trimToEmpty(input.getValue()));
  }

  public void setValueDouble(double val) {
    input.setValue(Double.toString(val));
  }

  private double safeParseDouble(String value) {
    try {
      if (value != null && value.length() > 0) {
        return Double.parseDouble(value);
      }
    } catch (NumberFormatException e) {
    }
    return 0.0d;
  }

  public FormInputText setTitle(String title) {
    this.title = title;
    return this;
  }

  public FormInputText setAutocomplete(String autocomplete) {
    this.autocomplete = autocomplete;
    return this;
  }

  public FormInputText setHasDanger(boolean value) {
    StyleUtils.toggleClass(rootEl, "has-danger", value);
    StyleUtils.hasClassName(rootEl, "has-danger");
    return this;
  }

  public boolean getHasDanger() {
    return StyleUtils.hasClassName(rootEl, "has-danger");
  }

  public FormInputText setType(String type) {
    this.type = type;
    return this;
  }

  public FormInputText setPlaceholder(String placeholder) {
    this.placeholder = placeholder;
    return this;
  }

  public FormInputText setRequired(boolean required) {
    this.required = required;
    return this;
  }

  public FormInputText setSpellcheck(boolean spellcheck) {
    this.spellcheck = spellcheck;
    return this;
  }

  public FormInputText setName(String name) {
    this.name = name;
    return this;
  }

  public String getName() {
    return name;
  }

  public FormInputText setGrid(boolean b) {
    this.grid = b;
    return this;
  }

  public void removeControlFeedback() {
    if (feedbackEl != null) {
      if (rootEl.contains(feedbackEl)) {
        rootEl.removeChild(feedbackEl);
      }
    }
  }

  public void setControlFeedback(String message) {
    feedbackEl.setTextContent(message);
    rootEl.appendChild(feedbackEl);
  }

  @Override
  public void enterDocument() {
    super.enterDocument();
    
    if (required) {
      addHandlerRegistration(input.addEventListener(Event.FOCUSOUT, new EventListener() {
        @Override
        public void handleEvent(Event evt) {
          if (getValue().length() == 0) {
            setHasDanger(true);
            setControlFeedback("You can't leave this empty.");
          }
        }
      }, false));
      addHandlerRegistration(input.addEventListener(Event.FOCUSIN, new EventListener() {
        @Override
        public void handleEvent(Event evt) {
          if (getValue().length() == 0 && getHasDanger()) {
            setHasDanger(false);
            removeControlFeedback();
          }
        }
      }, false));
    }
  }
}
