package com.tracker.client.activities;

import com.tracker.client.controls.Component;

import elemental.json.JsonObject;

public interface SearchComponent {
  public interface Presenter {
    void clickSearch();
  }

  void setPresenter(Presenter presenter);

  JsonObject updateModel();

  void updateView(JsonObject a);

  Component getComponent();
}
