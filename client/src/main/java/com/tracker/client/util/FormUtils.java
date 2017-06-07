package com.tracker.client.util;

import java.util.ArrayList;
import java.util.List;

import com.tracker.client.activities.widgets.FormInputText;
import com.tracker.client.activities.widgets.PopupMessage;
import com.tracker.client.controls.Component;
import com.tracker.client.jso.ErrorJso;
import com.google.gwt.core.client.JsArray;

import elemental.json.JsonArray;
import elemental.json.JsonObject;

public class FormUtils {
  private Component component;

  public FormUtils(Component component) {
    this.component = component;
  }

  public List<FormInputText> getInputs() {
    List<FormInputText> inputs = new ArrayList<>();
    int cnt = component.getChildCount();
    for (int i = 0; i < cnt; i++) {
      Component child = component.getChildAt(i);
      if (child instanceof FormInputText) {
        inputs.add((FormInputText) child);
      }
    }
    return inputs;
  }

  public void clearErrors() {
    List<FormInputText> inputs = getInputs();
    for (FormInputText input : inputs) {
      input.removeControlFeedback();
      input.setHasDanger(false);
    }
  }

  public boolean updateViewErrors(JsArray<ErrorJso> errors) {
    clearErrors();

    if (errors != null && errors.length() > 0) {
      for (int i = 0; i < errors.length(); i++) {
        ErrorJso err = errors.get(i);
        List<FormInputText> inputs = getInputs();
        for (FormInputText input : inputs) {
          if (err.getField().equals(input.getName())) {
            input.setHasDanger(true); 
            input.setControlFeedback(err.getMessage());
            return true;
          }
        }
      }
    }
    return false;
  }

  public boolean updateViewErrors(JsonObject resp) {
    clearErrors();

    List<FormInputText> inputs = getInputs();
    JsonArray errors = resp.getArray("errors");
    if (errors != null && errors.length() > 0) {
      for (int i = 0; i < errors.length(); i++) {
        JsonObject err = errors.getObject(i);
        String field = err.getString("field");
        String message = err.getString("message");
        for (FormInputText input : inputs) {
          if (field.equals(input.getName())) {
            input.setHasDanger(true); 
            input.setControlFeedback(message);
            return true;
          }
        }
        component.addChild(new PopupMessage(message), true);
      }
      return true;
    }
    return false;
  }
}
